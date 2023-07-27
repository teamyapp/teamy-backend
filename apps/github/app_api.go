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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	teamyClient "github.com/teamyapp/teamy-backend/core/client"

	cloudProto "github.com/teamyapp/cloud/app/api/proto"
	cloudClient "github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/entity"
	"github.com/teamyapp/teamy-backend/apps/github/client"
	githubEntity "github.com/teamyapp/teamy-backend/apps/github/entity"
	appsProto "github.com/teamyapp/teamy-backend/apps/proto"
	"github.com/teamyapp/teamy-backend/core/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const githubAppPathPrefix = "/apps/github"
const teamIDParam = "teamId"

const codeReviewMaxWait = 24 * time.Hour
const authProvider = "github"

const pullRequestIconURL = "/assets/apps/pull_request_dark_green.svg"
const pullRequestIconHoverURL = "/assets/apps/pull_request_light_green.svg"

var taskIDWithTaskLinkPattern = regexp.MustCompile(`\[\(task:([\d]+)\)\]\(.+\)`)
var taskIDWithTaskLinkFormat = "[(task:%d)](%s)"
var taskIDWithBodyFormat = "%v\n[(task:%d)](%s)"
var taskLinkFormat = "%v/teams/%v/tasks/%v"

type AppAPI struct {
	config                                   AppConfig
	logger                                   telemetry.Logger
	cloudClientRegistry                      *cloudClient.Registry
	teamyClientRegistry                      *teamyClient.Registry
	githubAppInstallStateDao                 dao.GithubAppInstallState
	githubAppInstallationDao                 dao.GithubAppInstallation
	githubPullRequestDao                     dao.GithubPullRequest
	githubCodeReviewDao                      dao.GithubCodeReview
	githubRequiredUserActionDao              dao.GithubRequiredUserAction
	githubPullRequestInternalTaskRelationDao dao.GithubPullRequestInternalTaskRelation
	githubApp                                *client.GithubApp
	githubGraphQLAPI                         client.GraphQLAPI
	githubRESTAPI                            client.RESTAPI
	appsProto.UnimplementedGithubServer
	teamyWebUIBaseURL string
}

var _ runner.Service = (*AppAPI)(nil)
var _ appsProto.GithubServer = (*AppAPI)(nil)

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
	rn.WithGRPCServer(func(server *grpc.Server) {
		appsProto.RegisterGithubServer(server, a)
	})
	return nil
}

func (a AppAPI) webInstall(writer http.ResponseWriter, request *http.Request) {
	// Verify request sender is team owner
	ct := request.Context()
	teamIDRaw := chi.URLParam(request, teamIDParam)
	teamID, err := strconv.ParseUint(teamIDRaw, 10, 64)
	if err != nil {
		internalErr := errs.NewError(errs.InvalidArgument, "must provide teamId")
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	query := request.URL.Query()
	redirectURL := query.Get("redirectUrl")
	if len(redirectURL) == 0 {
		internalErr := errs.NewError(errs.InvalidArgument, "must provide redirectUrl")
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	genStateIDReq := &cloudProto.GenerateUniqueNumberRequest{SequenceName: "githubInstallationStateID"}
	genStateIDRes, err := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(request.Context(), genStateIDReq)
	if err != nil {
		internalErr := errs.FromGRPCErr(err)
		a.logger.ErrorWithContext(ct, internalErr)
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
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	installURL, internalErr := a.getInstallGithubAppURL(ct, state.ID)
	if internalErr != nil {
		a.logger.ErrorWithContext(ct, internalErr)
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
		internalErr := errs.NewError(errs.InvalidArgument, "fail to parse state ID")
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	rawInstallationID := query.Get("installation_id")
	if len(rawInstallationID) == 0 {
		internalErr := errs.NewError(errs.InvalidArgument, "must provide installation_id")
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	installationID64, err := strconv.ParseInt(rawInstallationID, 10, 32)
	if err != nil {
		internalErr := errs.NewError(errs.InvalidArgument, "installation_id must be int")
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	installationID := int(installationID64)
	state, internalErr := a.githubAppInstallStateDao.FindStateByID(ct, stateID)
	if internalErr != nil {
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	expireAt := state.CreatedAt.Add(a.config.InstallationValidDuration)
	now := time.Now().UTC()
	if expireAt.Before(now) {
		internalErr := errs.NewError(errs.InvalidOperation, "install app session expired")
		a.logger.ErrorWithContext(ct, internalErr)
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
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	internalErr = a.githubAppInstallStateDao.DeleteState(ct, stateID)
	if internalErr != nil {
		a.logger.ErrorWithContext(ct, internalErr)
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
		internalErr := errs.NewError(errs.InvalidArgument, "signature header must have 2 parts")
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	if bodySignatureHeaderParts[0] != "sha256" {
		internalErr := errs.NewError(errs.InvalidArgument, "signature header must start with sha256")
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	buf, err := io.ReadAll(request.Body)
	if err != nil {
		internalErr := errs.NewError(errs.IO, "fail to read request payload")
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	signature, err := hex.DecodeString(bodySignatureHeaderParts[1])
	if err != nil {
		internalErr := errs.NewError(errs.InvalidArgument, "fail to decode request body signature")
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	if !validateHMACSignature(buf, []byte(a.config.WebhookSecret), signature) {
		internalErr := errs.NewError(errs.InvalidArgument, fmt.Sprintf("invalid request body signature: signature=%v", bodySignatureHeaderParts[1]))
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	evtType := request.Header.Get("X-GitHub-Event")
	a.logger.InfoWithContext(ct, fmt.Sprintf("received event: deliveryID=%v EventType=%v", deliveryID, evtType))
	internalErr := a.processEvent(ct, githubEntity.EventType(evtType), buf)
	if internalErr != nil {
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

func (a AppAPI) webListRequiredActionsForCurrentUser(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := errs.NewError(errs.Unauthenticated, "user id not found")
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	teamIDRaw := chi.URLParam(request, teamIDParam)
	teamID, err := strconv.ParseUint(teamIDRaw, 10, 64)
	if err != nil {
		internalErr := errs.NewError(errs.InvalidArgument, "must provide teamId")
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	requiredUserActions, internalErr := a.githubRequiredUserActionDao.
		FindRequiredUserActionsByActionUserID(ct, teamID, userID)
	if internalErr != nil {
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	// TODO: receive notification from cloud and update required action status
	requiredUserActions, internalErr = a.refreshRequiredActionsStatus(ct, userID, requiredUserActions)
	if err != nil {
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	requiredUserActions = collect.Filter(requiredUserActions, func(action entity.GithubRequiredUserAction) bool {
		return action.IsCompleted == false
	})
	web.WriteJSONToResponse(writer, requiredUserActions)
}

func (a AppAPI) webCreateRequiredAction(writer http.ResponseWriter, request *http.Request) {
	ct := request.Context()
	requestSenderID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := errs.NewError(errs.Unauthenticated, "user id not found")
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	teamIDRaw := chi.URLParam(request, teamIDParam)
	teamID, err := strconv.ParseUint(teamIDRaw, 10, 64)
	if err != nil {
		internalErr := errs.NewError(errs.InvalidArgument, "must provide teamId")
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	body := struct {
		UserActionType entity.GithubUserActionType `json:"userActionType"`
		ActionUserID   uint64                      `json:"actionUserId"`
	}{}
	buf, err := io.ReadAll(request.Body)
	if err != nil {
		internalErr := errs.NewError(errs.IO, err.Error())
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	err = json.Unmarshal(buf, &body)
	if err != nil {
		internalErr := errs.NewError(errs.Deserialization, err.Error())
		a.logger.ErrorWithContext(ct, internalErr)
		errs.SetHTTPErr(internalErr, writer)
		return
	}

	genActionIDReq := &cloudProto.GenerateUniqueNumberRequest{SequenceName: "githubRequiredActionID"}
	genActionIDRes, err := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genActionIDReq)
	if err != nil {
		internalErr := errs.FromGRPCErr(err)
		a.logger.ErrorWithContext(ct, internalErr)
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
		a.logger.ErrorWithContext(ct, internalErr)
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
			return entity.GithubRequiredUserAction{}, internalErr
		}
	}

	return requiredAction, nil
}

func (a AppAPI) processEvent(ct context.Context, evtType githubEntity.EventType, payload []byte) *errs.Error {
	var evt githubEntity.Event
	err := json.Unmarshal(payload, &evt)
	if err != nil {
		internalErr := errs.NewError(errs.Deserialization, err.Error())
		return internalErr
	}

	ins, internalErr := a.githubAppInstallationDao.FindInstallationByID(ct, evt.Installation.ID)
	if internalErr != nil {
		return internalErr
	}

	switch evtType {
	case githubEntity.PullRequestEventType:
		return a.processPullRequestEvent(ct, ins.TeamID, evt, payload)
	case githubEntity.PullRequestReviewEventType:
		return a.processPullRequestReviewEvent(ct, ins.TeamID, evt, payload)
	default:
		a.logger.InfoWithContext(ct, fmt.Sprintf("unknown event: eventType=%v", evtType))
	}

	return nil
}

func (a AppAPI) processPullRequestEvent(ct context.Context, teamID uint64, evt githubEntity.Event, payload []byte) *errs.Error {
	if evt.Sender.Type == githubEntity.OrganizationAccountType {
		return errs.NewError(errs.InvalidArgument, fmt.Sprintf("unsupported sender type: senderType=%v", evt.Sender.Type))
	}

	// https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#pull_request
	var prEvt githubEntity.PullRequestEvent
	err := json.Unmarshal(payload, &prEvt)
	if err != nil {
		return errs.NewError(errs.Deserialization, err.Error())
	}

	switch prEvt.Action {
	case githubEntity.OpenedPullRequestAction, githubEntity.ReopenedPullRequestAction:
		return a.createTaskForPullRequest(ct, teamID, evt, prEvt)
	case githubEntity.EditedPullRequestAction:
		if evt.Sender.Type != entity.GithubSenderTypeBot {
			return a.updateTaskForPullRequest(ct, teamID, evt, prEvt)
		}
	case githubEntity.AssignedPullRequestAction:
	case githubEntity.ReviewRequestedPullRequestAction:
		return a.createTaskForRequestedReviewers(ct, teamID, evt, prEvt)
	case githubEntity.ReviewRequestRemovedPullRequestAction:
	case githubEntity.ConvertedToDraftPullRequestAction:
	case githubEntity.ClosedPullRequestAction:
		if prEvt.PullRequest.Merged {
			return a.movePullRequestToDelivered(ct, prEvt)
		} else {
			return a.closePullRequest(ct, teamID, prEvt)
		}
	}

	return nil
}

func (a AppAPI) movePullRequestToDelivered(ct context.Context, prEvt githubEntity.PullRequestEvent) *errs.Error {
	prTaskRelations, err := a.githubPullRequestInternalTaskRelationDao.FindPullRequestInternalTaskRelationsByNodeID(ct, prEvt.PullRequest.NodeID)
	if err != nil {
		return err
	}

	for _, prTaskRelation := range prTaskRelations {
		a.moveTaskToDelivered(ct, prTaskRelation.InternalTaskID)
		if prTaskRelation.AutomaticTracking {
			a.deleteNonDeliveredWaitForTasks(ct, prTaskRelation.InternalTaskID)
		}
	}

	return nil
}

func (a AppAPI) moveTaskToDelivered(ct context.Context, taskID uint64) *errs.Error {
	moveTaskToDeliveredRequest := &proto.MoveTaskToDeliveredRequest{
		TaskId: taskID,
	}
	_, rpcErr := a.teamyClientRegistry.TaskClient().MoveTaskToDelivered(ct, moveTaskToDeliveredRequest)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return internalErr
	}

	return nil
}

func (a AppAPI) closePullRequest(ct context.Context, teamID uint64, prEvt githubEntity.PullRequestEvent) *errs.Error {
	prTaskRelations, err := a.githubPullRequestInternalTaskRelationDao.FindPullRequestInternalTaskRelationsByNodeID(ct, prEvt.PullRequest.NodeID)
	if err != nil {
		return err
	}

	if len(prTaskRelations) > 0 {
		body := prEvt.PullRequest.Body
		for _, prTaskRelation := range prTaskRelations {
			if prTaskRelation.AutomaticTracking {
				a.deleteNonDeliveredWaitForTasks(ct, prTaskRelation.InternalTaskID)
			}

			err := a.RemovePullRequestTaskRelationAndCleanup(ct, prTaskRelation)
			if err != nil {
				return err
			}

			taskIDWithTaskURL := a.formatTaskIDWithTaskURL(teamID, prTaskRelation.InternalTaskID)
			body = strings.ReplaceAll(body, taskIDWithTaskURL, "")
		}

		githubAppInstallation, err := a.githubAppInstallationDao.FindInstallationByTeamID(ct, teamID)
		if err != nil {
			return err
		}

		installation := a.githubApp.GetInstallation(githubAppInstallation.ID)

		_, err = a.githubGraphQLAPI.UpdatePullRequest(ct, installation, client.UpdatePullRequestInput{
			PullRequestID: prEvt.PullRequest.NodeID,
			Body:          &body,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (a AppAPI) deleteNonDeliveredWaitForTasks(ct context.Context, awaitingTaskID uint64) *errs.Error {
	getAwaitForTasksRequest := &proto.GetAwaitForTasksRequest{
		AwaitingTaskId: awaitingTaskID,
	}

	getAwaitForTasksResponse, rpcErr := a.teamyClientRegistry.TaskClient().GetAwaitForTasks(ct, getAwaitForTasksRequest)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return internalErr
	}

	for _, awaitForTask := range getAwaitForTasksResponse.Tasks {
		if awaitForTask.Status != proto.TaskStatus_Delivered {
			_, rpcErr = a.teamyClientRegistry.TaskClient().DeleteTask(ct, &proto.DeleteTaskRequest{
				TaskId: awaitForTask.TaskId,
			})
			if rpcErr != nil {
				internalErr := errs.FromGRPCErr(rpcErr)
				return internalErr
			}
		}
	}

	return nil
}

func (a AppAPI) tryGetValidMentionedTasks(ct context.Context, body string) (map[uint64]*proto.TaskMsg, *errs.Error) {
	tasks := map[uint64]*proto.TaskMsg{}
	allMatches := taskIDWithTaskLinkPattern.FindAllStringSubmatch(body, -1)
	for _, matches := range allMatches {
		taskID, err := strconv.ParseUint(matches[1], 10, 64)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		getTaskReq := &proto.GetTaskRequest{
			TaskId: taskID,
		}
		task, rpcErr := a.teamyClientRegistry.TaskClient().GetTask(ct, getTaskReq)
		if rpcErr != nil {
			internalErr := errs.FromGRPCErr(rpcErr)
			if internalErr.Code == errs.NotFound {
				a.logger.WarningWithContext(ct,
					internalErr.String(),
				)
			} else {
				return tasks, internalErr
			}
		} else {
			tasks[task.TaskId] = task
		}
	}

	return tasks, nil
}

func (a AppAPI) moveTaskToInProgress(ct context.Context, taskID uint64, teamID uint64) *errs.Error {
	moveTaskToInProgressReq := &proto.MoveTaskToInProgressRequest{TaskId: taskID}
	_, rpcErr := a.teamyClientRegistry.TaskClient().MoveTaskToInProgress(ct, moveTaskToInProgressReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return internalErr
	}

	a.logger.InfoWithContext(ct, fmt.Sprintf("task moved to in progress: taskID=%v", taskID))
	a.tryAddTaskToCurrentSprint(ct, teamID, taskID)
	return nil
}

func (a AppAPI) createPullRequestTaskRelation(
	ct context.Context,
	taskID uint64,
	automaticTracking bool,
	pullRequestURL string,
	pullRequestNodeID string,
) *errs.Error {
	iconURL := pullRequestIconURL
	iconHoverURL := pullRequestIconHoverURL
	createTaskLinkReq := &proto.CreateTaskLinkRequest{
		TaskId:       taskID,
		Title:        "View pull request on Github",
		Url:          pullRequestURL,
		IconUrl:      &iconURL,
		IconHoverUrl: &iconHoverURL,
	}

	createTaskLinkRes, rpcErr := a.teamyClientRegistry.TaskLinkClient().CreateTaskLink(ct, createTaskLinkReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return internalErr
	}

	pullRequestInternalTaskRelation := entity.GithubPullRequestInternalTaskRelation{
		PullRequestNodeID:  pullRequestNodeID,
		InternalTaskID:     taskID,
		InternalTaskLinkID: createTaskLinkRes.LinkId,
		AutomaticTracking:  automaticTracking,
	}

	return a.githubPullRequestInternalTaskRelationDao.CreatePullRequestInternalTaskRelation(
		ct, pullRequestInternalTaskRelation,
	)
}

func (a AppAPI) RemovePullRequestTaskRelationAndCleanup(ct context.Context, prTaskRelation entity.GithubPullRequestInternalTaskRelation) *errs.Error {
	if prTaskRelation.AutomaticTracking {
		deleteTaskReq := &proto.DeleteTaskRequest{
			TaskId: prTaskRelation.InternalTaskID,
		}

		_, rpcErr := a.teamyClientRegistry.TaskClient().DeleteTask(ct, deleteTaskReq)
		if rpcErr != nil {
			internalErr := errs.FromGRPCErr(rpcErr)
			return internalErr
		}
	} else {
		deleteTaskLinkReq := &proto.DeleteTaskLinkRequest{
			LinkId: prTaskRelation.InternalTaskLinkID,
		}

		_, rpcErr := a.teamyClientRegistry.TaskLinkClient().DeleteTaskLink(ct, deleteTaskLinkReq)
		if rpcErr != nil {
			internalErr := errs.FromGRPCErr(rpcErr)
			return internalErr
		}
	}

	err := a.githubPullRequestInternalTaskRelationDao.DeletePullRequestInternalTaskRelationByNodeIDAndTaskID(ct, prTaskRelation.PullRequestNodeID, prTaskRelation.InternalTaskID)
	if err != nil {
		return err
	}

	return nil
}

func (a AppAPI) removePullRequestTaskRelationsByTaskID(ct context.Context, installation *client.Installation, teamID uint64, taskID uint64) *errs.Error {
	prTaskRelations, err := a.githubPullRequestInternalTaskRelationDao.FindPullRequestInternalTaskRelationsByInternalTaskID(ct, taskID)
	if err != nil {
		return err
	}

	for _, prTaskRelation := range prTaskRelations {
		deleteTaskLinkReq := &proto.DeleteTaskLinkRequest{
			LinkId: prTaskRelation.InternalTaskLinkID,
		}

		_, rpcErr := a.teamyClientRegistry.TaskLinkClient().DeleteTaskLink(ct, deleteTaskLinkReq)
		if rpcErr != nil {
			internalErr := errs.FromGRPCErr(rpcErr)
			return internalErr
		}

		err = a.githubPullRequestInternalTaskRelationDao.DeletePullRequestInternalTaskRelationByNodeIDAndTaskID(ct, prTaskRelation.PullRequestNodeID, prTaskRelation.InternalTaskID)
		if err != nil {
			return err
		}

		remainingPrTaskRelations, err := a.githubPullRequestInternalTaskRelationDao.FindPullRequestInternalTaskRelationsByNodeID(ct, prTaskRelation.PullRequestNodeID)
		if err != nil {
			return err
		}

		pullRequestNode, err := a.githubGraphQLAPI.GetPullRequestByNodeID(ct, installation, prTaskRelation.PullRequestNodeID)
		if err != nil {
			return err
		}

		var body string
		if len(remainingPrTaskRelations) == 0 {
			taskID, err := a.createAutomaticTrackingTask(
				ct,
				teamID,
				pullRequestNode.Repository.Name,
				pullRequestNode.Author.ID,
				pullRequestNode.Number,
				pullRequestNode.Title,
				pullRequestNode.Body,
				pullRequestNode.URL,
				prTaskRelation.PullRequestNodeID,
			)
			if err != nil {
				return err
			}

			prevTaskIDWithTaskURL := a.formatTaskIDWithTaskURL(teamID, prTaskRelation.InternalTaskID)
			newTaskIDWithTaskURL := a.formatTaskIDWithTaskURL(teamID, *taskID)
			body = strings.ReplaceAll(
				pullRequestNode.Body,
				prevTaskIDWithTaskURL,
				newTaskIDWithTaskURL,
			)
		} else {
			taskIDWithTaskLink := a.formatTaskIDWithTaskURL(teamID, prTaskRelation.InternalTaskID)
			body = strings.ReplaceAll(
				pullRequestNode.Body,
				taskIDWithTaskLink,
				"",
			)
		}

		_, err = a.githubGraphQLAPI.UpdatePullRequest(ct, installation, client.UpdatePullRequestInput{
			Body:          &body,
			PullRequestID: prTaskRelation.PullRequestNodeID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (a AppAPI) createTaskForPullRequest(ct context.Context, teamID uint64, evt githubEntity.Event, prEvt githubEntity.PullRequestEvent) *errs.Error {
	pr := entity.GithubPullRequest{
		NodeID:          prEvt.PullRequest.NodeID,
		RepositoryOwner: &evt.Repository.Owner.Login,
		RepositoryName:  &evt.Repository.Name,
		Number:          &prEvt.Number,
		URL:             &prEvt.PullRequest.HtmlURL,
	}
	err := a.githubPullRequestDao.CreatePullRequest(ct, pr)
	if err != nil {
		return err
	}

	mentionedTasks, err := a.tryGetValidMentionedTasks(ct, prEvt.PullRequest.Body)
	if err != nil {
		return err
	}

	githubAppInstallation, err := a.githubAppInstallationDao.FindInstallationByTeamID(ct, teamID)
	if err != nil {
		return err
	}

	installation := a.githubApp.GetInstallation(githubAppInstallation.ID)
	if len(mentionedTasks) == 0 {
		taskID, err := a.createAutomaticTrackingTask(
			ct,
			teamID,
			evt.Repository.Name,
			prEvt.PullRequest.User.NodeID,
			prEvt.Number,
			prEvt.PullRequest.Title,
			prEvt.PullRequest.Body,
			*pr.URL,
			pr.NodeID)
		if err != nil {
			return err
		}

		body := a.formatTaskIDWithBody(prEvt.PullRequest.Body, teamID, *taskID)
		_, err = a.githubGraphQLAPI.UpdatePullRequest(ct, installation, client.UpdatePullRequestInput{
			PullRequestID: prEvt.PullRequest.NodeID,
			Body:          &body,
		})
		if err != nil {
			return err
		}
	} else {
		for _, task := range mentionedTasks {
			err = a.removePullRequestTaskRelationsByTaskID(ct, installation, teamID, task.TaskId)
			if err != nil {
				return err
			}

			err = a.createPullRequestTaskRelation(ct, task.TaskId, false, *pr.URL, pr.NodeID)
			if err != nil {
				return err
			}

			err = a.moveTaskToInProgress(ct, task.TaskId, teamID)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func (a AppAPI) updateTaskForPullRequest(ct context.Context, teamID uint64, evt githubEntity.Event, prEvt githubEntity.PullRequestEvent) *errs.Error {
	githubAppInstallation, err := a.githubAppInstallationDao.FindInstallationByTeamID(ct, teamID)
	if err != nil {
		return err
	}

	installation := a.githubApp.GetInstallation(githubAppInstallation.ID)
	prTaskRelations, err := a.githubPullRequestInternalTaskRelationDao.FindPullRequestInternalTaskRelationsByNodeID(ct, prEvt.PullRequest.NodeID)
	if err != nil {
		return err
	}

	prTaskRelationMap := map[uint64]entity.GithubPullRequestInternalTaskRelation{}
	for _, prTaskRelation := range prTaskRelations {
		prTaskRelationMap[prTaskRelation.InternalTaskID] = prTaskRelation
	}

	mentionedTasks, err := a.tryGetValidMentionedTasks(ct, prEvt.PullRequest.Body)
	if err != nil {
		return err
	}

	for _, task := range mentionedTasks {
		prTaskRelation, ok := prTaskRelationMap[task.TaskId]
		if !ok {
			err = a.removePullRequestTaskRelationsByTaskID(ct, installation, teamID, task.TaskId)
			if err != nil {
				return err
			}

			err = a.createPullRequestTaskRelation(
				ct,
				task.TaskId,
				false,
				prEvt.PullRequest.HtmlURL,
				prEvt.PullRequest.NodeID)
			if err != nil {
				return err
			}

			switch prEvt.PullRequest.State {
			case githubEntity.OpenPullRequestState:
				{
					err = a.moveTaskToInProgress(ct, task.TaskId, teamID)
					if err != nil {
						return err
					}
				}
			case githubEntity.ClosedPullRequestState:
				{
					if prEvt.PullRequest.Merged {
						err = a.moveTaskToDelivered(ct, task.TaskId)
						if err != nil {
							return err
						}
					}
				}
			}

		} else if prTaskRelation.AutomaticTracking {
			updateTaskReq := &proto.UpdateTaskRequest{
				TaskId:       task.TaskId,
				OwningTeamId: teamID,
				Goal:         fmt.Sprintf("[%v][PR #%v] %v", evt.Repository.Name, prEvt.Number, prEvt.PullRequest.Title),
				Context:      &prEvt.PullRequest.Body,
				OwnerUserId:  task.OwnerUserId,
				Effort:       task.Effort,
				DueAt:        task.DueAt,
			}

			_, rpcErr := a.teamyClientRegistry.TaskClient().UpdateTask(ct, updateTaskReq)
			if rpcErr != nil {
				internalErr := errs.FromGRPCErr(rpcErr)
				return internalErr
			}
		}
	}

	for _, prTaskRelation := range prTaskRelations {
		_, ok := mentionedTasks[prTaskRelation.InternalTaskID]
		if !ok {
			err := a.RemovePullRequestTaskRelationAndCleanup(ct, prTaskRelation)
			if err != nil {
				return err
			}
		}
	}

	if len(mentionedTasks) == 0 {
		taskID, err := a.createAutomaticTrackingTask(
			ct,
			teamID,
			evt.Repository.Name,
			prEvt.PullRequest.User.NodeID,
			prEvt.PullRequest.Number,
			prEvt.PullRequest.Title,
			prEvt.PullRequest.Body,
			prEvt.PullRequest.HtmlURL,
			prEvt.PullRequest.NodeID)
		if err != nil {
			return err
		}

		body := a.formatTaskIDWithBody(prEvt.PullRequest.Body, teamID, *taskID)
		_, err = a.githubGraphQLAPI.UpdatePullRequest(ct, installation, client.UpdatePullRequestInput{
			PullRequestID: prEvt.PullRequest.NodeID,
			Body:          &body,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (a AppAPI) createAutomaticTrackingTask(
	ct context.Context,
	teamID uint64,
	repositoryName string,
	pullRequestUserNodeID string,
	pullRequestNumber int,
	pullRequestTitle string,
	pullRequestBody string,
	pullRequestURL string,
	pullRequestNodeID string,
) (*uint64, *errs.Error) {
	prAuthorUserID, err := a.GetInternalUserID(ct, pullRequestUserNodeID)
	if err != nil {
		return nil, err
	}

	createTaskReq := &proto.CreateTaskRequest{
		TeamId:      teamID,
		Goal:        fmt.Sprintf("[%v][PR #%v] %v", repositoryName, pullRequestNumber, pullRequestTitle),
		Context:     &pullRequestBody,
		OwnerUserId: &prAuthorUserID,
	}
	createTaskRes, rpcErr := a.teamyClientRegistry.TaskClient().CreateTask(ct, createTaskReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		a.logger.LogWithContext(ct, telemetry.Error, telemetry.Props{
			telemetry.CauseProp: internalErr,
		})
		return nil, internalErr
	}

	a.logger.LogWithContext(ct, telemetry.Info, telemetry.Props{
		telemetry.MessageProp: fmt.Sprintf("pull request task created: repo=%v prNumber=%v taskID=%v",
			repositoryName,
			pullRequestNumber,
			createTaskRes.TaskId),
	})

	err = a.createPullRequestTaskRelation(
		ct,
		createTaskRes.TaskId,
		true,
		pullRequestURL,
		pullRequestNodeID)
	if err != nil {
		return nil, err
	}

	err = a.moveTaskToInProgress(ct, createTaskRes.TaskId, teamID)
	if err != nil {
		return nil, err
	}

	return &createTaskRes.TaskId, nil
}

func (a AppAPI) processPullRequestReviewEvent(ct context.Context, teamID uint64, evt githubEntity.Event, payload []byte) *errs.Error {
	if evt.Sender.Type == githubEntity.OrganizationAccountType {
		return errs.NewError(errs.InvalidArgument, fmt.Sprintf("unsupported sender type: senderType=%v", evt.Sender.Type))
	}

	// https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#pull_request
	var prReviewEvt githubEntity.PullRequestReviewEvent
	err := json.Unmarshal(payload, &prReviewEvt)
	if err != nil {
		return errs.NewError(errs.Deserialization, "failed to deserialize pull request review event")
	}

	codeReview, internalErr := a.githubCodeReviewDao.FindCodeReviewByGithubReviewerID(ct, prReviewEvt.PullRequest.NodeID, prReviewEvt.Review.User.NodeID)
	if internalErr != nil {
		return internalErr
	}

	internalErr = a.processGithubCodeReviewFeedback(ct, teamID, codeReview, evt, prReviewEvt)
	if internalErr != nil {
		return internalErr
	}

	a.logger.InfoWithContext(ct, fmt.Sprintf("moved review task to delivered: taskID=%v", codeReview.InternalCodeReviewTaskID))
	moveTaskToDeliveredRequest := &proto.MoveTaskToDeliveredRequest{
		TaskId: codeReview.InternalCodeReviewTaskID,
	}
	_, rpcErr := a.teamyClientRegistry.TaskClient().MoveTaskToDelivered(ct, moveTaskToDeliveredRequest)
	if rpcErr != nil {
		internalErr = errs.FromGRPCErr(rpcErr)
		return internalErr
	}

	return nil
}

func (a AppAPI) processGithubCodeReviewFeedback(ct context.Context, teamID uint64, codeReview entity.GithubCodeReview, evt githubEntity.Event, prReviewEvt githubEntity.PullRequestReviewEvent) *errs.Error {
	switch prReviewEvt.Action {
	case githubEntity.SubmittedPullRequestReviewAction:
		switch prReviewEvt.Review.State {
		case githubEntity.CommentedPullRequestReviewState, githubEntity.ChangesRequestedPullRequestReviewState:
			prAuthorUserID, err := a.GetInternalUserID(ct, prReviewEvt.PullRequest.User.NodeID)
			if err != nil {
				return err
			}

			prReviewerID, err := a.GetInternalUserID(ct, prReviewEvt.PullRequest.User.NodeID)
			if err != nil {
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
				return err
			}

			addressFeedbackTaskID := createTaskRes.TaskId
			a.logger.InfoWithContext(ct, fmt.Sprintf("address feedback task created: repo=%v, prNumber=%v, taskID=%v",
				evt.Repository.Name,
				prReviewEvt.PullRequest.Number,
				createTaskRes.TaskId))
			pr, err := a.githubPullRequestDao.FindPullRequestByGithubNodeID(ct, prReviewEvt.PullRequest.NodeID)
			if err != nil {
				return err
			}

			prTaskRelations, err := a.githubPullRequestInternalTaskRelationDao.FindPullRequestInternalTaskRelationsByNodeID(ct, pr.NodeID)
			if err != nil {
				return err
			}

			for _, prTaskRelation := range prTaskRelations {
				addAwaitForTaskReq := &proto.AddAwaitForTaskRequest{
					AwaitingTaskId: prTaskRelation.InternalTaskID,
					AwaitForTaskId: addressFeedbackTaskID,
				}
				_, rpcErr = a.teamyClientRegistry.TaskClient().AddAwaitForTask(ct, addAwaitForTaskReq)
				if rpcErr != nil {
					err = errs.FromGRPCErr(rpcErr)
					return err
				}
			}

			a.logger.InfoWithContext(ct, fmt.Sprintf("pull request is waiting for address feedback task: repo=%v, prNumber=%v, taskID=%v",
				evt.Repository.Name,
				prReviewEvt.PullRequest.Number,
				createTaskRes.TaskId))
			codeReview.InternalAddressFeedbackTaskID = &addressFeedbackTaskID
			return a.githubCodeReviewDao.UpdateCodeReview(ct, codeReview)
		case githubEntity.ApprovedPullRequestReviewState:
			// TODO: create merge task to wait for CI pipeline
		}
	}

	return nil
}

func (a AppAPI) GetInternalUserID(ct context.Context, githubUserNodeID string) (uint64, *errs.Error) {
	getInternalUserIdReq := &cloudProto.GetInternalUserIdRequest{AuthProvider: "github", ExternalUserId: githubUserNodeID}
	getInternalUserIdRes, rpcErr := a.cloudClientRegistry.IdentityClient().GetInternalUserId(ct, getInternalUserIdReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return 0, internalErr
	}

	return getInternalUserIdRes.InternalUserId, nil
}

func (a AppAPI) BackfillPullRequestMetadata(ct context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	var err *errs.Error
	pullRequests, err := a.githubPullRequestDao.FindAllPullRequests(ct)
	if err != nil {
		return nil, errs.ToGRPCErr(err)
	}

	pullRequests = collect.Filter(pullRequests, func(pullRequest entity.GithubPullRequest) bool {
		return pullRequest.RepositoryName == nil ||
			pullRequest.RepositoryOwner == nil ||
			pullRequest.URL == nil ||
			pullRequest.Number == nil ||
			pullRequest.OrganizationID == nil
	})

	for _, pullRequest := range pullRequests {
		a.logger.InfoWithContext(
			ct,
			fmt.Sprintf("start backfilling pull request, taskID=%d, nodeID=%s",
				pullRequest.InternalTaskID,
				pullRequest.NodeID,
			),
		)

		getTaskReq := &proto.GetTaskRequest{TaskId: pullRequest.InternalTaskID}
		task, rpcErr := a.teamyClientRegistry.TaskClient().GetTask(ct, getTaskReq)
		if rpcErr != nil {
			if err == nil {
				err = errs.FromGRPCErr(rpcErr)
			}

			a.logger.ErrorWithContext(ct, errs.FromGRPCErr(rpcErr))
			continue
		}

		installationID, sqlErr := a.githubAppInstallationDao.FindInstallationIDByTeamID(ct, task.OwningTeamId)
		if sqlErr != nil {
			if err == nil {
				err = sqlErr
			}

			a.logger.ErrorWithContext(ct, sqlErr)
			continue
		}

		ins := a.githubApp.GetInstallation(installationID)
		node, gqlErr := a.githubGraphQLAPI.GetPullRequestByNodeID(ct, ins, pullRequest.NodeID)
		if gqlErr != nil {
			if err == nil {
				err = gqlErr
			}

			a.logger.ErrorWithContext(ct, gqlErr)
			continue
		}

		org, err := a.githubRESTAPI.GetOrganizationByLogin(ct, ins, node.Repository.Owner.Login)
		if err != nil {
			a.logger.ErrorWithContext(ct, err)
			continue
		}

		gpr := entity.GithubPullRequest{
			InternalTaskID:  pullRequest.InternalTaskID,
			NodeID:          pullRequest.NodeID,
			RepositoryOwner: &node.Repository.Owner.Login,
			RepositoryName:  &node.Repository.Name,
			Number:          &node.Number,
			URL:             &node.URL,
			OrganizationID:  &org.ID,
		}

		sqlErr = a.githubPullRequestDao.UpdatePullRequest(ct, gpr)
		if sqlErr != nil {
			if err == nil {
				err = sqlErr
			}

			a.logger.ErrorWithContext(ct, sqlErr)
			continue
		}

		a.logger.InfoWithContext(
			ct,
			fmt.Sprintf("finish backfilling pull request, metadata=%v", gpr.String()))
	}

	return &emptypb.Empty{}, errs.ToGRPCErr(err)
}

func (a AppAPI) createTaskForRequestedReviewers(ct context.Context, teamID uint64, evt githubEntity.Event, prEvt githubEntity.PullRequestEvent) *errs.Error {
	pr, err := a.githubPullRequestDao.FindPullRequestByGithubNodeID(ct, prEvt.PullRequest.NodeID)
	if err != nil {
		return err
	}

	for _, githubReviewer := range prEvt.PullRequest.RequestedReviewers {
		prTaskRelations, err := a.githubPullRequestInternalTaskRelationDao.FindPullRequestInternalTaskRelationsByNodeID(ct, pr.NodeID)
		if err != nil {
			return err
		}

		for _, prTaskRelation := range prTaskRelations {
			err = a.tryCreateTaskForPullRequestReviewer(ct, teamID, prEvt.PullRequest.NodeID, prTaskRelation.InternalTaskID, githubReviewer.NodeID, evt, prEvt)
			if err != nil {
				a.logger.ErrorWithContext(ct, err)
				continue
			}

		}
	}

	return nil
}

func (a AppAPI) tryCreateTaskForPullRequestReviewer(
	ct context.Context,
	teamID uint64,
	githubPullRequestNodeID string,
	pullRequestTaskID uint64,
	githubReviewerNodeID string,
	evt githubEntity.Event,
	prEvt githubEntity.PullRequestEvent,
) *errs.Error {
	codeReview, err := a.githubCodeReviewDao.FindCodeReviewByGithubReviewerID(ct, githubPullRequestNodeID, githubReviewerNodeID)
	if err != nil && err.Code != errs.NotFound {
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
		createdTaskID, err := a.createCodeReviewTask(ct, teamID, pullRequestTaskID, githubReviewerNodeID, codeReview.Round+1, evt, prEvt)
		if err != nil {
			return err
		}

		codeReview.InternalCodeReviewTaskID = createdTaskID
		codeReview.InternalAddressFeedbackTaskID = nil
		codeReview.Round++
		_, rpcErr := a.teamyClientRegistry.TaskClient().MoveTaskToDelivered(ct, moveTaskToDeliveredRequest)
		if rpcErr != nil {
			internalErr := errs.FromGRPCErr(rpcErr)
			return internalErr
		}

		return a.githubCodeReviewDao.UpdateCodeReview(ct, codeReview)
	}

	createdTaskID, err := a.createCodeReviewTask(ct, teamID, pullRequestTaskID, githubReviewerNodeID, 1, evt, prEvt)
	if err != nil {
		return err
	}

	codeReview = entity.GithubCodeReview{
		GithubPullRequestNodeID:  githubPullRequestNodeID,
		InternalCodeReviewTaskID: createdTaskID,
		GithubReviewerNodeID:     githubReviewerNodeID,
		Round:                    1,
	}

	err = a.githubCodeReviewDao.CreateCodeReview(ct, codeReview)
	if err != nil {
		return err
	}

	return a.tryAddTaskToCurrentSprint(ct, teamID, createdTaskID)
}

func (a AppAPI) createCodeReviewTask(ct context.Context, teamID uint64, pullRequestTaskID uint64, githubReviewerNodeID string, round int, evt githubEntity.Event, prEvt githubEntity.PullRequestEvent) (uint64, *errs.Error) {
	codeReviewerInternalUserID, err := a.GetInternalUserID(ct, githubReviewerNodeID)
	if err != nil {
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
		return 0, internalErr
	}

	a.logger.InfoWithContext(ct, fmt.Sprintf("review task created: repo=%v, prNumber=%v, taskID=%v",
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
		return 0, internalErr
	}

	addAwaitForTaskReq := &proto.AddAwaitForTaskRequest{
		AwaitingTaskId: pullRequestTaskID,
		AwaitForTaskId: createTaskRes.TaskId,
	}

	_, rpcErr = a.teamyClientRegistry.TaskClient().AddAwaitForTask(ct, addAwaitForTaskReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return 0, internalErr
	}

	a.logger.InfoWithContext(ct, fmt.Sprintf("pull request is waiting for review task: prTaskID=%v, GithubReviewerID=%v, reviewTaskID=%v",
		pullRequestTaskID,
		githubReviewerNodeID,
		createTaskRes.TaskId))
	return createTaskRes.TaskId, nil
}

func (a AppAPI) tryAddTaskToCurrentSprint(ct context.Context, teamID uint64, taskID uint64) *errs.Error {
	getCurrentSprintReq := &proto.GetActiveSprintRequest{TeamId: teamID}
	getCurrentSprintRes, rpcErr := a.teamyClientRegistry.SprintClient().GetActiveSprint(ct, getCurrentSprintReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		if internalErr.Code == errs.NotFound {
			return nil
		}

		return internalErr
	}

	addTaskToSprintReq := &proto.AddTaskToSprintRequest{TaskId: taskID, SprintId: getCurrentSprintRes.Id}
	_, rpcErr = a.teamyClientRegistry.SprintClient().AddTaskToSprint(ct, addTaskToSprintReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return internalErr
	}

	return nil
}

func (a AppAPI) getInstallGithubAppURL(ct context.Context, stateID uint64) (string, *errs.Error) {
	urlStr := fmt.Sprintf("https://github.com/apps/%s/installations/new", a.config.AppName)
	installURL, err := url.Parse(urlStr)
	if err != nil {
		return "", errs.NewError(errs.Unknown, fmt.Sprintf("fail to parse URL: url=%v", urlStr))
	}

	query := url.Values{}
	query.Set("state", strconv.FormatUint(stateID, 10))
	installURL.RawQuery = query.Encode()
	return installURL.String(), nil
}

func (a AppAPI) formatTaskIDWithTaskURL(teamID uint64, taskID uint64) string {
	taskURL := a.formatTeamyWebTaskURL(teamID, taskID)
	return fmt.Sprintf(taskIDWithTaskLinkFormat, taskID, taskURL)
}

func (a AppAPI) formatTaskIDWithBody(body string, teamID uint64, taskID uint64) string {
	taskURL := a.formatTeamyWebTaskURL(teamID, taskID)
	return fmt.Sprintf(taskIDWithBodyFormat, body, taskID, taskURL)
}

func (a AppAPI) formatTeamyWebTaskURL(teamID uint64, taskID uint64) string {
	return fmt.Sprintf(taskLinkFormat, a.teamyWebUIBaseURL, teamID, taskID)
}

func NewAppAPI(
	cfg AppConfig,
	logger telemetry.Logger,
	cloudClientRegistry *cloudClient.Registry,
	teamyClientRegistry *teamyClient.Registry,
	githubAppInstallStateDao dao.GithubAppInstallState,
	githubAppInstallationDao dao.GithubAppInstallation,
	githubPullRequestDao dao.GithubPullRequest,
	githubCodeReviewDao dao.GithubCodeReview,
	githubRequiredUserActionDao dao.GithubRequiredUserAction,
	githubPullRequestInternalTaskRelationDao dao.GithubPullRequestInternalTaskRelation,
	githubGraphQLAPI client.GraphQLAPI,
	githubRESTAPI client.RESTAPI,
	githubApp *client.GithubApp,
	teamyWebUIBaseURL string,
) AppAPI {

	return AppAPI{
		config:                                   cfg,
		logger:                                   logger,
		cloudClientRegistry:                      cloudClientRegistry,
		teamyClientRegistry:                      teamyClientRegistry,
		githubAppInstallStateDao:                 githubAppInstallStateDao,
		githubAppInstallationDao:                 githubAppInstallationDao,
		githubPullRequestDao:                     githubPullRequestDao,
		githubCodeReviewDao:                      githubCodeReviewDao,
		githubRequiredUserActionDao:              githubRequiredUserActionDao,
		githubPullRequestInternalTaskRelationDao: githubPullRequestInternalTaskRelationDao,
		githubGraphQLAPI:                         githubGraphQLAPI,
		githubRESTAPI:                            githubRESTAPI,
		githubApp:                                githubApp,
		teamyWebUIBaseURL:                        teamyWebUIBaseURL,
	}
}

func validateHMACSignature(message []byte, key []byte, signature []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(expectedMAC, signature)
}
