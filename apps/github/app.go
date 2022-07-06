package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	cloudProto "github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/entity"
	"github.com/teamyapp/teamy-backend/core/api"
)

const githubAppPathPrefix = "/apps/github"

type App struct {
	config                   AppConfig
	cloudClientRegistry      *cloudAPI.ClientRegistry
	teamyClientRegistry      *api.ClientRegistry
	githubAppInstallStateDao dao.GithubAppInstallState
	githubAppInstallationDao dao.GithubAppInstallation
}

var _ runner.Service = (*App)(nil)

func (a App) Start(rn *runner.ServiceRunner) error {
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Path:        path.Join(githubAppPathPrefix, "install"),
			Method:      http.MethodGet,
			HandlerFunc: a.install,
		},
		{
			Path:        path.Join(githubAppPathPrefix, "install", "finish"),
			Method:      http.MethodGet,
			HandlerFunc: a.finishInstall,
		},
		{
			Path:        path.Join(githubAppPathPrefix, "webhook"),
			Method:      http.MethodPost,
			HandlerFunc: a.onEventNotify,
		},
	})
	return nil
}

func (a App) install(w http.ResponseWriter, r *http.Request) {
	// Verify request sender is team owner
	query := r.URL.Query()
	teamID, err := strconv.ParseUint(query.Get("team-id"), 10, 64)
	if err != nil {
		log.Println("must provide team-id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	redirectURL := query.Get("redirect-url")
	if len(redirectURL) == 0 {
		log.Println("must provide redirect-url")
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

func (a App) finishInstall(w http.ResponseWriter, r *http.Request) {
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

func (a App) onEventNotify(w http.ResponseWriter, r *http.Request) {
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
	eventType := r.Header.Get("X-GitHub-Event")
	log.Printf("Github event received: deliveryID=%s, event=%s\n", deliveryID, eventType)
	err = a.processEvent(eventType, buf)
	if err != nil {
		log.Printf("fail to process Github event: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a App) processEvent(eventType string, payload []byte) error {
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

	switch eventType {
	case "pull_request":
		return a.processPullRequestEvent(ins, payload)
	default:
		log.Printf("Unknown event: %v\n", eventType)
	}

	return nil
}

func (a App) processPullRequestEvent(installation entity.GithubAppInstallation, payload []byte) error {
	// https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#pull_request
	var evt pullRequestEvent
	err := json.Unmarshal(payload, &evt)
	if err != nil {
		log.Println(err)
		return err
	}

	switch evt.Action {
	case openedPullRequestAction:
	case assignedPullRequestAction:
	case reviewRequestedPullRequestAction:
	case reviewRequestRemovedPullRequestAction:
	case convertedToDraftPullRequestAction:
	case closedPullRequestAction:
	}
	return nil
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
) App {
	return App{
		config:                   cfg,
		cloudClientRegistry:      cloudClientRegistry,
		teamyClientRegistry:      teamyClientRegistry,
		githubAppInstallStateDao: githubAppInstallStateDao,
		githubAppInstallationDao: githubAppInstallationDao,
	}
}

func validateHMACSignature(message []byte, key []byte, signature []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(expectedMAC, signature)
}
