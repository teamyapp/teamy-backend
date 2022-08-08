package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	cloudAPI "github.com/teamyapp/cloud/app/api"
	cloudProto "github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/web"
	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/entity"
	"github.com/teamyapp/teamy-backend/core/api"
	"github.com/teamyapp/teamy-backend/core/api/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const githubAppPathPrefix = "/apps/github"
const codeReviewMaxWait = 24 * time.Hour

type App struct {
	config                      AppConfig
	cloudClientRegistry         *cloudAPI.ClientRegistry
	teamyClientRegistry         *api.ClientRegistry
	githubAppInstallStateDao    dao.GithubAppInstallState
	githubAppInstallationDao    dao.GithubAppInstallation
	githubPullRequestDao        dao.GithubPullRequest
	githubCodeReviewDao         dao.GithubCodeReview
	githubRequiredUserActionDao dao.GithubRequiredUserAction
}

var _ runner.Service = (*App)(nil)

func (a App) Start(rn *runner.ServiceRunner) error {
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Path:        path.Join(githubAppPathPrefix, "teams", "{teamId}", "install"),
			Method:      http.MethodGet,
			HandlerFunc: a.webInstall,
		},
		{
			Path:        path.Join(githubAppPathPrefix, "install", "finish"),
			Method:      http.MethodGet,
			HandlerFunc: a.webFinishInstall,
		},
		{
			Path:        path.Join(githubAppPathPrefix, "webhook"),
			Method:      http.MethodPost,
			HandlerFunc: a.webOnEventNotify,
		},
		{
			Path:        path.Join(githubAppPathPrefix, "teams", "{teamId}", "required-actions", "current-user"),
			Method:      http.MethodGet,
			HandlerFunc: a.webListRequiredActionsForCurrentUser,
		},
		{
			Path:        path.Join(githubAppPathPrefix, "teams", "{teamId}", "required-actions", "create"),
			Method:      http.MethodPost,
			HandlerFunc: a.webCreateRequiredAction,
		},
	})
	return nil
}

func (a App) webInstall(w http.ResponseWriter, r *http.Request) {
	// Verify request sender is team owner
	teamIDParam := mux.Vars(r)["teamId"]
	teamID, err := strconv.ParseUint(teamIDParam, 10, 64)
	if err != nil {
		log.Println("must provide teamId")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	query := r.URL.Query()
	redirectURL := query.Get("redirectUrl")
	if len(redirectURL) == 0 {
		log.Println("must provide redirectUrl")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	genStateIDReq := &cloudProto.GenerateUniqueNumberRequest{SequenceName: "githubInstallationStateID"}
	genStateIDRes, err := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(context.Background(), genStateIDReq)
	if err != nil {
		log.Printf("fail to generate state ID: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	state := entity.GithubAppInstallState{
		ID:          genStateIDRes.UniqueNumber,
		TeamID:      teamID,
		RedirectURL: redirectURL,
		CreatedAt:   time.Now().UTC(),
	}
	err = a.githubAppInstallStateDao.CreateState(state)
	if err != nil {
		log.Printf("fail to create state: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	installURL, err := a.getInstallGithubAppURL(state.ID)
	if err != nil {
		log.Printf("fail to get Github App install URL: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, installURL, http.StatusTemporaryRedirect)
}

func (a App) webFinishInstall(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	stateIDParam := query.Get("state")
	stateID, err := strconv.ParseUint(stateIDParam, 10, 64)
	if err != nil {
		log.Printf("fail to parse state ID: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	state, err := a.githubAppInstallStateDao.FindStateByID(stateID)
	if err != nil {
		log.Printf("fail to find state ID: state ID=%v, err=%v\n", stateID, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	expireAt := state.CreatedAt.Add(a.config.InstallationValidDuration)
	now := time.Now().UTC()
	if expireAt.Before(now) {
		log.Println("install app session expired")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	installationID := query.Get("installation_id")
	ins := entity.GithubAppInstallation{
		ID:        installationID,
		TeamID:    state.TeamID,
		CreatedAt: time.Now().UTC(),
	}
	err = a.githubAppInstallationDao.CreateGithubAppInstallation(ins)
	if err != nil {
		log.Printf("fail to create Github App installation: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = a.githubAppInstallStateDao.DeleteState(stateID)
	if err != nil {
		log.Printf("fail to delete state: stateID=%v, err=%v\n", stateID, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, state.RedirectURL, http.StatusTemporaryRedirect)
}

func (a App) webOnEventNotify(w http.ResponseWriter, r *http.Request) {
	bodySignatureHeader := r.Header.Get("X-Hub-Signature-256")
	bodySignatureHeaderParts := strings.Split(bodySignatureHeader, "=")
	if len(bodySignatureHeaderParts) != 2 {
		log.Println("invalid signature format")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if bodySignatureHeaderParts[0] != "sha256" {
		log.Println("signature header mus start with sha256")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	buf, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("fail to read request payload: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	signature, err := hex.DecodeString(bodySignatureHeaderParts[1])
	if err != nil {
		log.Printf("fail to decode request body signature: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !validateHMACSignature(buf, []byte(a.config.WebhookSecret), signature) {
		log.Printf("invalid request body signature: signature=%v\n", bodySignatureHeaderParts[1])
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	deliveryID := r.Header.Get("X-GitHub-Delivery")
	evtType := r.Header.Get("X-GitHub-Event")
	log.Printf("Github event received: deliveryID=%s, event=%s\n", deliveryID, evtType)
	ct, _ := context.WithTimeout(context.Background(), a.config.RequestTimeout)
	err = a.processEvent(ct, eventType(evtType), buf)
	if err != nil {
		log.Printf("fail to process Github event: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a App) webListRequiredActionsForCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, err := ctx.UserIDFromContext(r.Context())
	if err != nil {
		log.Println("must provide userId")
		w.WriteHeader(http.StatusUnauthorized)
	}

	teamIDParam := mux.Vars(r)["teamId"]
	teamID, err := strconv.ParseUint(teamIDParam, 10, 64)
	if err != nil {
		log.Println("must provide teamId")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	requiredUserActions, err := a.githubRequiredUserActionDao.FindRequiredUserActionsByUserID(teamID, userID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO: receive notification from cloud and update required action status
	requiredUserActions, err = a.refreshRequiredActionsStatus(requiredUserActions)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	requiredUserActions = collect.Filter(requiredUserActions, func(action entity.GithubRequiredUserAction) bool {
		return action.IsCompleted == false
	})
	web.WriteJSON(w, requiredUserActions)
}

func (a App) webCreateRequiredAction(w http.ResponseWriter, r *http.Request) {
	requestSenderID, err := ctx.UserIDFromContext(r.Context())
	if err != nil {
		log.Println("must provide request sender ID")
		w.WriteHeader(http.StatusUnauthorized)
	}

	teamIDParam := mux.Vars(r)["teamId"]
	teamID, err := strconv.ParseUint(teamIDParam, 10, 64)
	if err != nil {
		log.Println("must provide teamId")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body := struct {
		UserActionType entity.GithubUserActionType `json:"userActionType"`
		ActionUserID   uint64                      `json:"actionUserId"`
	}{}
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(buf, &body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	genActionIDReq := &cloudProto.GenerateUniqueNumberRequest{SequenceName: "githubRequiredActionID"}
	genActionIDRes, err := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(context.Background(), genActionIDReq)
	if err != nil {
		log.Printf("fail to generate required action ID: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
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

	err = a.githubRequiredUserActionDao.CreateRequiredUserAction(action)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a App) refreshRequiredActionsStatus(
	requiredActions []entity.GithubRequiredUserAction,
) ([]entity.GithubRequiredUserAction, error) {
	// TODO: refresh required action status
	return requiredActions, nil
}

func (a App) processEvent(ct context.Context, evtType eventType, payload []byte) error {
	// TODO: parse & react to Github event
	var evt event
	err := json.Unmarshal(payload, &evt)
	if err != nil {
		log.Println(err)
		return err
	}

	ins, err := a.githubAppInstallationDao.FindInstallationByID(evt.Installation.ID)
	if err != nil {
		log.Println(err)
		return err
	}

	switch evtType {
	case pullRequestEventType:
		return a.processPullRequestEvent(ct, ins.TeamID, evt, payload)
	case pullRequestReviewEventType:
		return a.processPullRequestReviewEvent(ct, ins.TeamID, evt, payload)
	default:
		log.Printf("Unknown event: %v\n", evtType)
	}

	return nil
}

func (a App) processPullRequestEvent(ct context.Context, teamID uint64, evt event, payload []byte) error {
	if evt.Sender.Type == organizationAccountType {
		return fmt.Errorf("unsupported account type: %v", evt.Sender.Type)
	}

	// https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#pull_request
	var prEvt pullRequestEvent
	err := json.Unmarshal(payload, &prEvt)
	if err != nil {
		log.Println(err)
		return err
	}

	switch prEvt.Action {
	case openedPullRequestAction, reopenedPullRequestAction:
		return a.createTaskForPullRequest(ct, teamID, evt, prEvt)
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

func (a App) movePullRequestToDelivered(ct context.Context, prEvt pullRequestEvent) error {
	pr, err := a.githubPullRequestDao.FindPullRequestByGithubNodeID(prEvt.PullRequest.NodeID)
	if err != nil {
		log.Println(err)
		return err
	}

	moveTaskToDeliveredRequest := &proto.MoveTaskToDeliveredRequest{
		TaskId: pr.InternalTaskID,
	}
	_, err = a.teamyClientRegistry.TaskClient().MoveTaskToDelivered(ct, moveTaskToDeliveredRequest)
	return err
}

func (a App) createTaskForPullRequest(ct context.Context, teamID uint64, evt event, prEvt pullRequestEvent) error {
	prAuthorUserID, err := a.GetInternalUserID(ct, prEvt.PullRequest.User.ID)
	if err != nil {
		log.Println(err)
		return err
	}

	createTaskReq := &proto.CreateTaskRequest{
		TeamId:      teamID,
		Goal:        fmt.Sprintf("[%v][PR #%v] %v", evt.Repository.Name, prEvt.Number, prEvt.PullRequest.Title),
		Context:     &prEvt.PullRequest.Body,
		OwnerUserId: &prAuthorUserID,
	}
	createTaskRes, err := a.teamyClientRegistry.TaskClient().CreateTask(ct, createTaskReq)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("Pull request task created: repo=%v pullRequestNumber=%v taskID=%v\n", evt.Repository.Name, prEvt.Number, createTaskRes.TaskId)
	moveTaskToInProgressReq := &proto.MoveTaskToInProgressRequest{TaskId: createTaskRes.TaskId}
	_, err = a.teamyClientRegistry.TaskClient().MoveTaskToInProgress(ct, moveTaskToInProgressReq)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("Task moved to in progress: taskID=%v\n", createTaskRes.TaskId)
	pr := entity.GithubPullRequest{
		NodeID:         prEvt.PullRequest.NodeID,
		InternalTaskID: createTaskRes.TaskId,
	}
	err = a.githubPullRequestDao.CreatePullRequest(pr)
	if err != nil {
		log.Println(err)
		return err
	}

	return a.tryAddTaskToCurrentSprint(ct, teamID, createTaskRes.TaskId)
}

func (a App) processPullRequestReviewEvent(ct context.Context, teamID uint64, evt event, payload []byte) error {
	if evt.Sender.Type == organizationAccountType {
		return fmt.Errorf("unsupported account type: %v", evt.Sender.Type)
	}

	// https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#pull_request
	var prReviewEvt pullRequestReviewEvent
	err := json.Unmarshal(payload, &prReviewEvt)
	if err != nil {
		log.Println(err)
		return err
	}

	codeReview, err := a.githubCodeReviewDao.FindCodeReviewByGithubReviewerID(prReviewEvt.PullRequest.NodeID, prReviewEvt.Review.User.ID)
	if err != nil {
		log.Println(err)
		return err
	}

	err = a.processGithubCodeReviewFeedback(ct, teamID, codeReview, evt, prReviewEvt)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("Moved review task to delivered: taskID=%v\n", codeReview.InternalCodeReviewTaskID)
	moveTaskToDeliveredRequest := &proto.MoveTaskToDeliveredRequest{
		TaskId: codeReview.InternalCodeReviewTaskID,
	}
	_, err = a.teamyClientRegistry.TaskClient().MoveTaskToDelivered(ct, moveTaskToDeliveredRequest)
	return err
}

func (a App) processGithubCodeReviewFeedback(ct context.Context, teamID uint64, codeReview entity.GithubCodeReview, evt event, prReviewEvt pullRequestReviewEvent) error {
	switch prReviewEvt.Action {
	case submittedPullRequestReviewAction:
		switch prReviewEvt.Review.State {
		case commentedPullRequestReviewState, changesRequestedPullRequestReviewState:
			prAuthorUserID, err := a.GetInternalUserID(ct, prReviewEvt.PullRequest.User.ID)
			if err != nil {
				log.Println(err)
				return err
			}

			prReviewerID, err := a.GetInternalUserID(ct, prReviewEvt.PullRequest.User.ID)
			if err != nil {
				log.Println(err)
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
			createTaskRes, err := a.teamyClientRegistry.TaskClient().CreateTask(ct, createTaskReq)
			if err != nil {
				log.Println(err)
				return err
			}

			addressFeedbackTaskID := createTaskRes.TaskId
			log.Printf(
				"Address feedback task created: repo=%v pullRequestNumber=%v taskID=%v\n",
				evt.Repository.Name,
				prReviewEvt.PullRequest.Number,
				createTaskRes.TaskId)
			pr, err := a.githubPullRequestDao.FindPullRequestByGithubNodeID(prReviewEvt.PullRequest.NodeID)
			if err != nil {
				log.Println(err)
				return err
			}

			addAwaitForTaskReq := &proto.AddAwaitForTaskRequest{
				AwaitingTaskId: pr.InternalTaskID,
				AwaitForTaskId: addressFeedbackTaskID,
			}
			_, err = a.teamyClientRegistry.TaskClient().AddAwaitForTask(ct, addAwaitForTaskReq)
			if err != nil {
				log.Println(err)
				return err
			}

			log.Printf(
				"Pull request is waiting for address feedback task: repo=%v pullRequestNumber=%v, addressFeedbackTaskID=%v\n",
				evt.Repository.Name,
				prReviewEvt.PullRequest.Number,
				addressFeedbackTaskID)
			codeReview.InternalAddressFeedbackTaskID = &addressFeedbackTaskID
			return a.githubCodeReviewDao.UpdateCodeReview(codeReview)
		case approvedPullRequestReviewState:
			// TODO: create merge task to wait for CI pipeline
		}
	}

	return nil
}

func (a App) GetInternalUserID(ct context.Context, githubUserID uint64) (uint64, error) {
	githubReviewerIDStr := strconv.FormatUint(githubUserID, 10)
	getInternalUserIdReq := &cloudProto.GetInternalUserIdRequest{AuthProvider: "github", ExternalUserId: githubReviewerIDStr}
	getInternalUserIdRes, err := a.cloudClientRegistry.IdentityClient().GetInternalUserId(ct, getInternalUserIdReq)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	return getInternalUserIdRes.InternalUserId, nil
}

func (a App) createTaskForRequestedReviewers(ct context.Context, teamID uint64, evt event, prEvt pullRequestEvent) error {
	pr, err := a.githubPullRequestDao.FindPullRequestByGithubNodeID(prEvt.PullRequest.NodeID)
	if err != nil {
		log.Println(err)
		return err
	}

	for _, githubReviewer := range prEvt.PullRequest.RequestedReviewers {
		err = a.tryCreateTaskForPullRequestReviewer(ct, teamID, prEvt.PullRequest.NodeID, pr.InternalTaskID, githubReviewer.ID, evt, prEvt)
		if err != nil {
			log.Println(err)
			continue
		}
	}

	return nil
}

func (a App) tryCreateTaskForPullRequestReviewer(
	ct context.Context,
	teamID uint64,
	githubPullRequestNodeID string,
	pullRequestTaskID uint64,
	githubReviewerID uint64,
	evt event,
	prEvt pullRequestEvent,
) error {
	codeReview, err := a.githubCodeReviewDao.FindCodeReviewByGithubReviewerID(githubPullRequestNodeID, githubReviewerID)
	var errNotFound dao.ErrNotFound
	if err != nil && !errors.As(err, &errNotFound) {
		log.Println(err)
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
			log.Println(err)
			return err
		}

		codeReview.InternalCodeReviewTaskID = createdTaskID
		codeReview.InternalAddressFeedbackTaskID = nil
		codeReview.Round++
		_, err = a.teamyClientRegistry.TaskClient().MoveTaskToDelivered(ct, moveTaskToDeliveredRequest)
		if err != nil {
			log.Println(err)
			return err
		}

		return a.githubCodeReviewDao.UpdateCodeReview(codeReview)
	}

	createdTaskID, err := a.createCodeReviewTask(ct, teamID, pullRequestTaskID, githubReviewerID, 1, evt, prEvt)
	if err != nil {
		log.Println(err)
		return err
	}

	codeReview = entity.GithubCodeReview{
		GithubPullRequestNodeID:  githubPullRequestNodeID,
		InternalCodeReviewTaskID: createdTaskID,
		GithubReviewerID:         githubReviewerID,
		Round:                    1,
	}

	err = a.githubCodeReviewDao.CreateCodeReview(codeReview)
	if err != nil {
		log.Println(err)
		return err
	}

	return a.tryAddTaskToCurrentSprint(ct, teamID, createdTaskID)
}

func (a App) createCodeReviewTask(ct context.Context, teamID uint64, pullRequestTaskID uint64, githubReviewerID uint64, round int, evt event, prEvt pullRequestEvent) (uint64, error) {
	codeReviewerInternalUserID, err := a.GetInternalUserID(ct, githubReviewerID)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	dueAt := time.Now().UTC().Add(codeReviewMaxWait)
	createTaskReq := &proto.CreateTaskRequest{
		TeamId:      teamID,
		Goal:        fmt.Sprintf("[%v][PR #%v] Code review round %v", evt.Repository.Name, prEvt.PullRequest.Number, round),
		OwnerUserId: &codeReviewerInternalUserID,
		DueAt:       timestamppb.New(dueAt),
	}
	createTaskRes, err := a.teamyClientRegistry.TaskClient().CreateTask(ct, createTaskReq)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	log.Printf("Review task created: repo=%v pullRequestNumber=%v taskID=%v\n", evt.Repository.Name, prEvt.PullRequest.Number, createTaskRes.TaskId)
	addAwaitForTaskReq := &proto.AddAwaitForTaskRequest{
		AwaitingTaskId: pullRequestTaskID,
		AwaitForTaskId: createTaskRes.TaskId,
	}

	_, err = a.teamyClientRegistry.TaskClient().AddAwaitForTask(ct, addAwaitForTaskReq)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	log.Printf("Pull request is waiting for review task: pullRequestTaskID=%v, githubReviewerID=%v, reviewTaskID=%v\n", pullRequestTaskID, githubReviewerID, createTaskRes.TaskId)
	return createTaskRes.TaskId, nil
}

func (a App) tryAddTaskToCurrentSprint(ct context.Context, teamID uint64, taskID uint64) error {
	getCurrentSprintReq := &proto.GetCurrentSprintRequest{TeamId: teamID}
	getCurrentSprintRes, err := a.teamyClientRegistry.SprintClient().GetCurrentSprint(ct, getCurrentSprintReq)
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.NotFound {
			return nil
		}

		log.Println(err)
		return err
	}

	addTaskToSprintReq := &proto.AddTaskToSprintRequest{TaskId: taskID, SprintId: getCurrentSprintRes.Id}
	_, err = a.teamyClientRegistry.SprintClient().AddTaskToSprint(ct, addTaskToSprintReq)
	return err
}

func (a App) getInstallGithubAppURL(stateID uint64) (string, error) {
	urlStr := fmt.Sprintf("https://github.com/apps/%s/installations/new", a.config.AppName)
	installURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	query := url.Values{}
	query.Set("state", strconv.FormatUint(stateID, 10))
	installURL.RawQuery = query.Encode()
	return installURL.String(), nil
}

func NewApp(
	cfg AppConfig,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	teamyClientRegistry *api.ClientRegistry,
	githubAppInstallStateDao dao.GithubAppInstallState,
	githubAppInstallationDao dao.GithubAppInstallation,
	githubPullRequestDao dao.GithubPullRequest,
	githubCodeReviewDao dao.GithubCodeReview,
	githubRequiredUserActionDao dao.GithubRequiredUserAction,
) App {
	return App{
		config:                      cfg,
		cloudClientRegistry:         cloudClientRegistry,
		teamyClientRegistry:         teamyClientRegistry,
		githubAppInstallStateDao:    githubAppInstallStateDao,
		githubAppInstallationDao:    githubAppInstallationDao,
		githubPullRequestDao:        githubPullRequestDao,
		githubCodeReviewDao:         githubCodeReviewDao,
		githubRequiredUserActionDao: githubRequiredUserActionDao,
	}
}

func validateHMACSignature(message []byte, key []byte, signature []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(expectedMAC, signature)
}
