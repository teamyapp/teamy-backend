package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	cloudAPI "github.com/teamyapp/cloud/app/api"
	cloudProto "github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/entity"
	"github.com/teamyapp/teamy-backend/apps/github/client"
	"github.com/teamyapp/teamy-backend/core/api"
	"github.com/teamyapp/teamy-backend/core/api/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const githubAppPathPrefix = "/apps/github"
const teamIDParam = "teamId"

const codeReviewMaxWait = 24 * time.Hour
const authProvider = "github"

const pullRequestIconURL = "/assets/apps/pull_request_dark_green.svg"
const pullRequestIconHoverURL = "/assets/apps/pull_request_light_green.svg"

type AppAPI struct {
	config                      AppConfig
	dataCollector               telemetry.DataCollector
	cloudClientRegistry         *cloudAPI.ClientRegistry
	teamyClientRegistry         *api.ClientRegistry
	githubAppInstallStateDao    dao.GithubAppInstallState
	githubAppInstallationDao    dao.GithubAppInstallation
	githubPullRequestDao        dao.GithubPullRequest
	githubCodeReviewDao         dao.GithubCodeReview
	githubRequiredUserActionDao dao.GithubRequiredUserAction
	githubApp                   *client.GithubApp
	githubGraphQLAPI            client.GraphQLAPI
}

var _ runner.Service = (*AppAPI)(nil)

func (a AppAPI) Start(rn *runner.ServiceRunner) *errs.Error {
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Method:      http.MethodGet,
			Pattern:     path.Join(githubAppPathPrefix, "teams", runner.Param(teamIDParam), "install"),
			HandlerFunc: a.webInstall,
		},
		{
			Method:      http.MethodGet,
			Pattern:     path.Join(githubAppPathPrefix, "install", "finish"),
			HandlerFunc: a.webFinishInstall,
		},
		{
			Pattern:     path.Join(githubAppPathPrefix, "webhook"),
			Method:      http.MethodPost,
			HandlerFunc: a.webOnEventNotify,
		},
		{
			Method:      http.MethodGet,
			Pattern:     path.Join(githubAppPathPrefix, "teams", runner.Param(teamIDParam), "required-actions", "current-user"),
			HandlerFunc: a.webListRequiredActionsForCurrentUser,
		},
		{
			Method:      http.MethodPost,
			Pattern:     path.Join(githubAppPathPrefix, "teams", runner.Param(teamIDParam), "required-actions", "create"),
			HandlerFunc: a.webCreateRequiredAction,
		},
	})
	return nil
}

func (a AppAPI) webInstall(writer http.ResponseWriter, request *http.Request) {
	// Verify request sender is team owner
	ct := request.Context()
	teamIDRaw := chi.URLParam(request, teamIDParam)
	teamID, err := strconv.ParseUint(teamIDRaw, 10, 64)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: err,
			Message:  "must provide teamId",
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	query := request.URL.Query()
	redirectURL := query.Get("redirectUrl")
	if len(redirectURL) == 0 {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: err,
			Message:  "must provide redirectUrl",
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	genStateIDReq := &cloudProto.GenerateUniqueNumberRequest{SequenceName: "githubInstallationStateID"}
	genStateIDRes, err := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(request.Context(), genStateIDReq)
	if err != nil {
		internalErr := errs.FromGRPCErr(err)
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	state := entity.GithubAppInstallState{
		ID:          genStateIDRes.UniqueNumber,
		TeamID:      teamID,
		RedirectURL: redirectURL,
		CreatedAt:   time.Now().UTC(),
	}
	internalErr := a.githubAppInstallStateDao.CreateState(ct, state)
	if internalErr != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	installURL, internalErr := a.getInstallGithubAppURL(ct, state.ID)
	if internalErr != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	http.Redirect(writer, request, installURL, http.StatusTemporaryRedirect)
}

func (a AppAPI) webFinishInstall(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	query := request.URL.Query()
	stateIDParam := query.Get("state")
	stateID, err := strconv.ParseUint(stateIDParam, 10, 64)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: err,
			Message:  "fail to parse state ID",
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	installationID := query.Get("installation_id")
	if len(installationID) == 0 {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: err,
			Message:  "must provide installation_id",
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	state, internalErr := a.githubAppInstallStateDao.FindStateByID(ct, stateID)
	if internalErr != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	expireAt := state.CreatedAt.Add(a.config.InstallationValidDuration)
	now := time.Now().UTC()
	if expireAt.Before(now) {
		internalErr = &errs.Error{
			Code:    errs.InvalidOperation,
			Message: "install app session expired",
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	ins := entity.GithubAppInstallation{
		ID:        installationID,
		TeamID:    state.TeamID,
		CreatedAt: time.Now().UTC(),
	}
	internalErr = a.githubAppInstallationDao.CreateGithubAppInstallation(ct, ins)
	if internalErr != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	internalErr = a.githubAppInstallStateDao.DeleteState(ct, stateID)
	if internalErr != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	http.Redirect(writer, request, state.RedirectURL, http.StatusTemporaryRedirect)
}

func (a AppAPI) webOnEventNotify(writer http.ResponseWriter, request *http.Request) {
	deliveryID := request.Header.Get("X-GitHub-Delivery")
	ct := request.Context()
	ct = ctx.NewContextWithRequestID(ct, deliveryID)

	bodySignatureHeader := request.Header.Get("X-Hub-Signature-256")
	bodySignatureHeaderParts := strings.Split(bodySignatureHeader, "=")
	if len(bodySignatureHeaderParts) != 2 {
		internalErr := &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "signature header must have 2 parts",
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	if bodySignatureHeaderParts[0] != "sha256" {
		internalErr := &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "signature header must start with sha256",
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	buf, err := io.ReadAll(request.Body)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.IO,
			EmbedErr: err,
			Message:  "fail to read request payload",
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	signature, err := hex.DecodeString(bodySignatureHeaderParts[1])
	if err != nil {
		internalErr := &errs.Error{
			Code:    errs.Unknown,
			Message: "fail to decode request body signature",
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	if !validateHMACSignature(buf, []byte(a.config.WebhookSecret), signature) {
		internalErr := &errs.Error{
			Code:    errs.InvalidArgument,
			Message: fmt.Sprintf("invalid request body signature: signature=%v", bodySignatureHeaderParts[1]),
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	evtType := request.Header.Get("X-GitHub-Event")
	a.dataCollector.Logger.InfoWithContext(ct, fmt.Sprintf("received event: deliveryID=%v EventType=%v", deliveryID, evtType))
	internalErr := a.processEvent(ct, eventType(evtType), buf)
	if internalErr != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (a AppAPI) webListRequiredActionsForCurrentUser(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user id not found",
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	teamIDRaw := chi.URLParam(request, teamIDParam)
	teamID, err := strconv.ParseUint(teamIDRaw, 10, 64)
	if err != nil {
		internalErr := &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "must provide teamId",
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	requiredUserActions, internalErr := a.githubRequiredUserActionDao.
		FindRequiredUserActionsByActionUserID(ct, teamID, userID)
	if internalErr != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	// TODO: receive notification from cloud and update required action status
	requiredUserActions, internalErr = a.refreshRequiredActionsStatus(ct, userID, requiredUserActions)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	requiredUserActions = collect.Filter(requiredUserActions, func(action entity.GithubRequiredUserAction) bool {
		return action.IsCompleted == false
	})
	web.WriteJSONToResponse(ct, a.dataCollector, writer, requiredUserActions)
}

func (a AppAPI) webCreateRequiredAction(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	requestSenderID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user id not found",
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	teamIDRaw := chi.URLParam(request, teamIDParam)
	teamID, err := strconv.ParseUint(teamIDRaw, 10, 64)
	if err != nil {
		internalErr := &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "must provide teamId",
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	body := struct {
		UserActionType entity.GithubUserActionType `json:"userActionType"`
		ActionUserID   uint64                      `json:"actionUserId"`
	}{}
	buf, err := io.ReadAll(request.Body)
	if err != nil {
		internalErr := &errs.Error{
			Code: errs.IO,
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	err = json.Unmarshal(buf, &body)
	if err != nil {
		internalErr := &errs.Error{
			Code: errs.Deserialization,
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	genActionIDReq := &cloudProto.GenerateUniqueNumberRequest{SequenceName: "githubRequiredActionID"}
	genActionIDRes, err := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genActionIDReq)
	if err != nil {
		internalErr := errs.FromGRPCErr(err)
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	action := entity.GithubRequiredUserAction{
		ID:                genActionIDRes.UniqueNumber,
		TeamID:            teamID,
		ActionUserID:      body.ActionUserID,
		UserActionType:    body.UserActionType,
		IsCompleted:       false,
		RequestedAt:       time.Now().UTC(),
		RequestedByUserID: requestSenderID,
	}

	internalErr := a.githubRequiredUserActionDao.CreateRequiredUserAction(ct, action)
	if internalErr != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (a AppAPI) refreshRequiredActionsStatus(
	ct context.Context,
	userID uint64,
	requiredActions []entity.GithubRequiredUserAction,
) ([]entity.GithubRequiredUserAction, *errs.Error) {
	refreshedRequiredActions := make([]entity.GithubRequiredUserAction, 0)
	for _, requiredAction := range requiredActions {
		if requiredAction.IsCompleted {
			refreshedRequiredActions = append(refreshedRequiredActions, requiredAction)
			continue
		}

		refreshedRequiredAction, err := a.refreshRequiredActionStatus(ct, userID, requiredAction)
		if err != nil {
			a.dataCollector.Logger.ErrorWithContext(ct, err)
			return nil, err
		}

		refreshedRequiredActions = append(refreshedRequiredActions, refreshedRequiredAction)
	}

	return refreshedRequiredActions, nil
}

func (a AppAPI) refreshRequiredActionStatus(
	ct context.Context,
	userID uint64,
	requiredAction entity.GithubRequiredUserAction,
) (entity.GithubRequiredUserAction, *errs.Error) {
	switch requiredAction.UserActionType {
	case entity.LinkGithubAccountGithubUserActionType:
		listUserLinksReq := &cloudProto.ListUserLinksRequest{InternalUserId: userID}
		listUserLinksRes, err := a.cloudClientRegistry.IdentityClient().ListUserLinks(ct, listUserLinksReq)
		if err != nil {
			internalErr := errs.FromGRPCErr(err)
			a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.GithubRequiredUserAction{}, internalErr
		}

		userLinks := collect.Filter(listUserLinksRes.UserLinks, func(userLink *cloudProto.UserLink) bool {
			return userLink.AuthProvider == authProvider
		})
		if len(userLinks) == 0 {
			return requiredAction, nil
		}

		requiredAction.IsCompleted = true
		internalErr := a.githubRequiredUserActionDao.UpdateRequiredUserAction(ct, requiredAction)
		if internalErr != nil {
			a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.GithubRequiredUserAction{}, internalErr
		}
	}

	return requiredAction, nil
}

func (a AppAPI) processEvent(ct context.Context, evtType eventType, payload []byte) *errs.Error {
	var evt event
	err := json.Unmarshal(payload, &evt)
	if err != nil {
		internalErr := &errs.Error{
			Code: errs.Deserialization,
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	ins, internalErr := a.githubAppInstallationDao.FindInstallationByID(ct, evt.Installation.ID)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	switch evtType {
	case pullRequestEventType:
		return a.processPullRequestEvent(ct, ins.TeamID, evt, payload)
	case pullRequestReviewEventType:
		return a.processPullRequestReviewEvent(ct, ins.TeamID, evt, payload)
	default:
		a.dataCollector.Logger.InfoWithContext(ct, fmt.Sprintf("unknown event: eventType=%v", evtType))
	}

	return nil
}

func (a AppAPI) processPullRequestEvent(ct context.Context, teamID uint64, evt event, payload []byte) *errs.Error {
	if evt.Sender.Type == organizationAccountType {
		internalErr := &errs.Error{
			Code:    errs.InvalidArgument,
			Message: fmt.Sprintf("unsupported sender type: senderType=%v", evt.Sender.Type),
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	// https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#pull_request
	var prEvt pullRequestEvent
	err := json.Unmarshal(payload, &prEvt)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Deserialization,
			EmbedErr: err,
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	switch prEvt.Action {
	case openedPullRequestAction, reopenedPullRequestAction:
		return a.createTaskForPullRequest(ct, teamID, evt, prEvt)
	case editedPullRequestAction:
		return a.updateTaskForPullRequest(ct, teamID, evt, prEvt)
	case assignedPullRequestAction:
	case reviewRequestedPullRequestAction:
		return a.createTaskForRequestedReviewers(ct, teamID, evt, prEvt)
	case reviewRequestRemovedPullRequestAction:
	case convertedToDraftPullRequestAction:
	case closedPullRequestAction:
		if prEvt.PullRequest.Merged {
			return a.movePullRequestToDelivered(ct, prEvt)
		}
	}

	return nil
}

func (a AppAPI) movePullRequestToDelivered(ct context.Context, prEvt pullRequestEvent) *errs.Error {
	pr, err := a.githubPullRequestDao.FindPullRequestByGithubNodeID(ct, prEvt.PullRequest.NodeID)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	moveTaskToDeliveredRequest := &proto.MoveTaskToDeliveredRequest{
		TaskId: pr.InternalTaskID,
	}
	_, rpcErr := a.teamyClientRegistry.TaskClient().MoveTaskToDelivered(ct, moveTaskToDeliveredRequest)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (a AppAPI) createTaskForPullRequest(ct context.Context, teamID uint64, evt event, prEvt pullRequestEvent) *errs.Error {
	prAuthorUserID, err := a.GetInternalUserID(ct, prEvt.PullRequest.User.ID)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	createTaskReq := &proto.CreateTaskRequest{
		TeamId:      teamID,
		Goal:        fmt.Sprintf("[%v][PR #%v] %v", evt.Repository.Name, prEvt.Number, prEvt.PullRequest.Title),
		Context:     &prEvt.PullRequest.Body,
		OwnerUserId: &prAuthorUserID,
	}
	createTaskRes, rpcErr := a.teamyClientRegistry.TaskClient().CreateTask(ct, createTaskReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	a.dataCollector.Logger.InfoWithContext(ct, fmt.Sprintf("pull request task created: repo=%v prNumber=%v taskID=%v",
		evt.Repository.Name,
		prEvt.Number,
		createTaskRes.TaskId))
	iconURL := pullRequestIconURL
	iconHoverURL := pullRequestIconHoverURL
	createTaskLinkReq := &proto.CreateTaskLinkRequest{
		TaskId:       createTaskRes.TaskId,
		Title:        "View pull request on Github",
		Url:          prEvt.PullRequest.HtmlURL,
		IconUrl:      &iconURL,
		IconHoverUrl: &iconHoverURL,
	}

	_, rpcErr = a.teamyClientRegistry.TaskLinkClient().CreateTaskLink(ct, createTaskLinkReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	moveTaskToInProgressReq := &proto.MoveTaskToInProgressRequest{TaskId: createTaskRes.TaskId}
	_, rpcErr = a.teamyClientRegistry.TaskClient().MoveTaskToInProgress(ct, moveTaskToInProgressReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	a.dataCollector.Logger.InfoWithContext(ct, fmt.Sprintf("task moved to in progress: taskID=%v", createTaskRes.TaskId))
	pr := entity.GithubPullRequest{
		NodeID:          prEvt.PullRequest.NodeID,
		InternalTaskID:  createTaskRes.TaskId,
		RepositoryOwner: evt.Repository.Owner.Login,
		RepositoryName:  evt.Repository.Name,
		Number:          prEvt.Number,
		URL:             prEvt.PullRequest.HtmlURL,
	}
	err = a.githubPullRequestDao.CreatePullRequest(ct, pr)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return a.tryAddTaskToCurrentSprint(ct, teamID, createTaskRes.TaskId)
}

func (a AppAPI) updateTaskForPullRequest(ct context.Context, teamID uint64, evt event, prEvt pullRequestEvent) *errs.Error {
	pr, err := a.githubPullRequestDao.FindPullRequestByGithubNodeID(ct, prEvt.PullRequest.NodeID)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	getTaskReq := &proto.GetTaskRequest{TaskId: pr.InternalTaskID}
	task, rpcErr := a.teamyClientRegistry.TaskClient().GetTask(ct, getTaskReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	updateTaskReq := &proto.UpdateTaskRequest{
		TaskId:       pr.InternalTaskID,
		Goal:         fmt.Sprintf("[%v][PR #%v] %v", evt.Repository.Name, prEvt.Number, prEvt.PullRequest.Title),
		Context:      &prEvt.PullRequest.Body,
		OwningTeamId: teamID,
		OwnerUserId:  task.OwnerUserId,
		Effort:       task.Effort,
		DueAt:        task.DueAt,
	}

	_, rpcErr = a.teamyClientRegistry.TaskClient().UpdateTask(ct, updateTaskReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (a AppAPI) processPullRequestReviewEvent(ct context.Context, teamID uint64, evt event, payload []byte) *errs.Error {
	if evt.Sender.Type == organizationAccountType {
		internalErr := &errs.Error{
			Code:    errs.InvalidArgument,
			Message: fmt.Sprintf("unsupported sender type: senderType=%v", evt.Sender.Type),
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	// https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#pull_request
	var prReviewEvt pullRequestReviewEvent
	err := json.Unmarshal(payload, &prReviewEvt)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Deserialization,
			EmbedErr: err,
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	codeReview, internalErr := a.githubCodeReviewDao.FindCodeReviewByGithubReviewerID(ct, prReviewEvt.PullRequest.NodeID, prReviewEvt.Review.User.ID)
	if internalErr != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	internalErr = a.processGithubCodeReviewFeedback(ct, teamID, codeReview, evt, prReviewEvt)
	if internalErr != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	a.dataCollector.Logger.InfoWithContext(ct, fmt.Sprintf("moved review task to delivered: taskID=%v", codeReview.InternalCodeReviewTaskID))
	moveTaskToDeliveredRequest := &proto.MoveTaskToDeliveredRequest{
		TaskId: codeReview.InternalCodeReviewTaskID,
	}
	_, rpcErr := a.teamyClientRegistry.TaskClient().MoveTaskToDelivered(ct, moveTaskToDeliveredRequest)
	if rpcErr != nil {
		internalErr = errs.FromGRPCErr(rpcErr)
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (a AppAPI) processGithubCodeReviewFeedback(ct context.Context, teamID uint64, codeReview entity.GithubCodeReview, evt event, prReviewEvt pullRequestReviewEvent) *errs.Error {
	switch prReviewEvt.Action {
	case submittedPullRequestReviewAction:
		switch prReviewEvt.Review.State {
		case commentedPullRequestReviewState, changesRequestedPullRequestReviewState:
			prAuthorUserID, err := a.GetInternalUserID(ct, prReviewEvt.PullRequest.User.ID)
			if err != nil {
				a.dataCollector.Logger.ErrorWithContext(ct, err)
				return err
			}

			prReviewerID, err := a.GetInternalUserID(ct, prReviewEvt.PullRequest.User.ID)
			if err != nil {
				a.dataCollector.Logger.ErrorWithContext(ct, err)
				return err
			}

			createTaskReq := &proto.CreateTaskRequest{
				TeamId: teamID,
				Goal: fmt.Sprintf(
					"[%v][PR #%v] Address round %v code review feedback from %v",
					evt.Repository.Name,
					prReviewEvt.PullRequest.Number,
					codeReview.Round,
					prReviewerID),
				Context:     &prReviewEvt.Review.Body,
				OwnerUserId: &prAuthorUserID,
			}
			createTaskRes, rpcErr := a.teamyClientRegistry.TaskClient().CreateTask(ct, createTaskReq)
			if rpcErr != nil {
				err = errs.FromGRPCErr(rpcErr)
				a.dataCollector.Logger.ErrorWithContext(ct, err)
				return err
			}

			addressFeedbackTaskID := createTaskRes.TaskId
			a.dataCollector.Logger.InfoWithContext(ct, fmt.Sprintf("address feedback task created: repo=%v, prNumber=%v, taskID=%v",
				evt.Repository.Name,
				prReviewEvt.PullRequest.Number,
				createTaskRes.TaskId))
			pr, err := a.githubPullRequestDao.FindPullRequestByGithubNodeID(ct, prReviewEvt.PullRequest.NodeID)
			if err != nil {
				a.dataCollector.Logger.ErrorWithContext(ct, err)
				return err
			}

			addAwaitForTaskReq := &proto.AddAwaitForTaskRequest{
				AwaitingTaskId: pr.InternalTaskID,
				AwaitForTaskId: addressFeedbackTaskID,
			}
			_, rpcErr = a.teamyClientRegistry.TaskClient().AddAwaitForTask(ct, addAwaitForTaskReq)
			if rpcErr != nil {
				err = errs.FromGRPCErr(rpcErr)
				a.dataCollector.Logger.ErrorWithContext(ct, err)
				return err
			}

			a.dataCollector.Logger.InfoWithContext(ct, fmt.Sprintf("pull request is waiting for address feedback task: repo=%v, prNumber=%v, taskID=%v",
				evt.Repository.Name,
				prReviewEvt.PullRequest.Number,
				createTaskRes.TaskId))
			codeReview.InternalAddressFeedbackTaskID = &addressFeedbackTaskID
			return a.githubCodeReviewDao.UpdateCodeReview(ct, codeReview)
		case approvedPullRequestReviewState:
			// TODO: create merge task to wait for CI pipeline
		}
	}

	return nil
}

func (a AppAPI) GetInternalUserID(ct context.Context, githubUserID uint64) (uint64, *errs.Error) {
	githubReviewerIDStr := strconv.FormatUint(githubUserID, 10)
	getInternalUserIdReq := &cloudProto.GetInternalUserIdRequest{AuthProvider: "github", ExternalUserId: githubReviewerIDStr}
	getInternalUserIdRes, rpcErr := a.cloudClientRegistry.IdentityClient().GetInternalUserId(ct, getInternalUserIdReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return 0, internalErr
	}

	return getInternalUserIdRes.InternalUserId, nil
}

func (a AppAPI) createTaskForRequestedReviewers(ct context.Context, teamID uint64, evt event, prEvt pullRequestEvent) *errs.Error {
	pr, err := a.githubPullRequestDao.FindPullRequestByGithubNodeID(ct, prEvt.PullRequest.NodeID)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	for _, githubReviewer := range prEvt.PullRequest.RequestedReviewers {
		err = a.tryCreateTaskForPullRequestReviewer(ct, teamID, prEvt.PullRequest.NodeID, pr.InternalTaskID, githubReviewer.ID, evt, prEvt)
		if err != nil {
			a.dataCollector.Logger.ErrorWithContext(ct, err)
			continue
		}
	}

	return nil
}

func (a AppAPI) tryCreateTaskForPullRequestReviewer(
	ct context.Context,
	teamID uint64,
	githubPullRequestNodeID string,
	pullRequestTaskID uint64,
	githubReviewerID uint64,
	evt event,
	prEvt pullRequestEvent,
) *errs.Error {
	codeReview, err := a.githubCodeReviewDao.FindCodeReviewByGithubReviewerID(ct, githubPullRequestNodeID, githubReviewerID)
	if err != nil && err.Code != errs.NotFound {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	codeReviewExist := err == nil
	if codeReviewExist {
		if codeReview.InternalAddressFeedbackTaskID == nil {
			// Discard repeated code review request
			return nil
		}

		// TODO: ensure author resolved at least 1 thread or pushed new commit
		// before feedback can be considered as addressed.
		moveTaskToDeliveredRequest := &proto.MoveTaskToDeliveredRequest{
			TaskId: *codeReview.InternalAddressFeedbackTaskID,
		}
		createdTaskID, err := a.createCodeReviewTask(ct, teamID, pullRequestTaskID, githubReviewerID, codeReview.Round+1, evt, prEvt)
		if err != nil {
			a.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		codeReview.InternalCodeReviewTaskID = createdTaskID
		codeReview.InternalAddressFeedbackTaskID = nil
		codeReview.Round++
		_, rpcErr := a.teamyClientRegistry.TaskClient().MoveTaskToDelivered(ct, moveTaskToDeliveredRequest)
		if rpcErr != nil {
			internalErr := errs.FromGRPCErr(rpcErr)
			a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return internalErr
		}

		return a.githubCodeReviewDao.UpdateCodeReview(ct, codeReview)
	}

	createdTaskID, err := a.createCodeReviewTask(ct, teamID, pullRequestTaskID, githubReviewerID, 1, evt, prEvt)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	codeReview = entity.GithubCodeReview{
		GithubPullRequestNodeID:  githubPullRequestNodeID,
		InternalCodeReviewTaskID: createdTaskID,
		GithubReviewerID:         githubReviewerID,
		Round:                    1,
	}

	err = a.githubCodeReviewDao.CreateCodeReview(ct, codeReview)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return a.tryAddTaskToCurrentSprint(ct, teamID, createdTaskID)
}

func (a AppAPI) createCodeReviewTask(ct context.Context, teamID uint64, pullRequestTaskID uint64, githubReviewerID uint64, round int, evt event, prEvt pullRequestEvent) (uint64, *errs.Error) {
	codeReviewerInternalUserID, err := a.GetInternalUserID(ct, githubReviewerID)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return 0, err
	}

	dueAt := time.Now().UTC().Add(codeReviewMaxWait)
	createTaskReq := &proto.CreateTaskRequest{
		TeamId:      teamID,
		Goal:        fmt.Sprintf("[%v][PR #%v] Code review round %v", evt.Repository.Name, prEvt.PullRequest.Number, round),
		OwnerUserId: &codeReviewerInternalUserID,
		DueAt:       timestamppb.New(dueAt),
	}
	createTaskRes, rpcErr := a.teamyClientRegistry.TaskClient().CreateTask(ct, createTaskReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return 0, internalErr
	}

	a.dataCollector.Logger.InfoWithContext(ct, fmt.Sprintf("review task created: repo=%v, prNumber=%v, taskID=%v",
		evt.Repository.Name,
		prEvt.PullRequest.Number,
		createTaskRes.TaskId))

	iconURL := pullRequestIconURL
	iconHoverURL := pullRequestIconHoverURL
	createTaskLinkReq := &proto.CreateTaskLinkRequest{
		TaskId:       createTaskRes.TaskId,
		Title:        "View pull request on Github",
		Url:          prEvt.PullRequest.HtmlURL,
		IconUrl:      &iconURL,
		IconHoverUrl: &iconHoverURL,
	}
	_, rpcErr = a.teamyClientRegistry.TaskLinkClient().CreateTaskLink(ct, createTaskLinkReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return 0, internalErr
	}

	addAwaitForTaskReq := &proto.AddAwaitForTaskRequest{
		AwaitingTaskId: pullRequestTaskID,
		AwaitForTaskId: createTaskRes.TaskId,
	}

	_, rpcErr = a.teamyClientRegistry.TaskClient().AddAwaitForTask(ct, addAwaitForTaskReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return 0, internalErr
	}

	a.dataCollector.Logger.InfoWithContext(ct, fmt.Sprintf("pull request is waiting for review task: prTaskID=%v, GithubReviewerID=%v, reviewTaskID=%v",
		pullRequestTaskID,
		githubReviewerID,
		createTaskRes.TaskId))
	return createTaskRes.TaskId, nil
}

func (a AppAPI) tryAddTaskToCurrentSprint(ct context.Context, teamID uint64, taskID uint64) *errs.Error {
	getCurrentSprintReq := &proto.GetCurrentSprintRequest{TeamId: teamID}
	getCurrentSprintRes, rpcErr := a.teamyClientRegistry.SprintClient().GetCurrentSprint(ct, getCurrentSprintReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		if internalErr.Code == errs.NotFound {
			return nil
		}

		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	addTaskToSprintReq := &proto.AddTaskToSprintRequest{TaskId: taskID, SprintId: getCurrentSprintRes.Id}
	_, rpcErr = a.teamyClientRegistry.SprintClient().AddTaskToSprint(ct, addTaskToSprintReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (a AppAPI) getInstallGithubAppURL(ct context.Context, stateID uint64) (string, *errs.Error) {
	urlStr := fmt.Sprintf("https://github.com/apps/%s/installations/new", a.config.AppName)
	installURL, err := url.Parse(urlStr)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
			Message:  fmt.Sprintf("fail to parse URL: url=%v", urlStr),
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", internalErr
	}

	query := url.Values{}
	query.Set("state", strconv.FormatUint(stateID, 10))
	installURL.RawQuery = query.Encode()
	return installURL.String(), nil
}

func NewAppAPI(
	cfg AppConfig,
	dataCollector telemetry.DataCollector,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	teamyClientRegistry *api.ClientRegistry,
	githubAppInstallStateDao dao.GithubAppInstallState,
	githubAppInstallationDao dao.GithubAppInstallation,
	githubPullRequestDao dao.GithubPullRequest,
	githubCodeReviewDao dao.GithubCodeReview,
	githubRequiredUserActionDao dao.GithubRequiredUserAction,
	githubGraphQLAPI client.GraphQLAPI,
	githubApp *client.GithubApp,
) AppAPI {
	return AppAPI{
		config:                      cfg,
		dataCollector:               dataCollector,
		cloudClientRegistry:         cloudClientRegistry,
		teamyClientRegistry:         teamyClientRegistry,
		githubAppInstallStateDao:    githubAppInstallStateDao,
		githubAppInstallationDao:    githubAppInstallationDao,
		githubPullRequestDao:        githubPullRequestDao,
		githubCodeReviewDao:         githubCodeReviewDao,
		githubRequiredUserActionDao: githubRequiredUserActionDao,
		githubGraphQLAPI:            githubGraphQLAPI,
		githubApp:                   githubApp,
	}
}

func validateHMACSignature(message []byte, key []byte, signature []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(expectedMAC, signature)
}
