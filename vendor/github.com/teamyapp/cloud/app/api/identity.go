package api

import (
	"io/ioutil"
	"net/http"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/web"
)

const identityPathPrefix = "/identity"

type Identity struct {
	identityService service.Identity
}

var _ runner.Service = (*Identity)(nil)

func (i Identity) Start(runner *runner.ServiceRunner) error {
	runner.RegisterWebRoutes([]web.Route{
		{
			Path:        path.Join(identityPathPrefix, "verify-token"),
			Method:      http.MethodPost,
			HandlerFunc: i.verifyToken,
		},
		{
			Path:        path.Join(identityPathPrefix, "sign-in/oauth/{provider}"),
			Method:      http.MethodGet,
			HandlerFunc: i.oauthSignIn,
		},
		{
			Path:        path.Join(identityPathPrefix, "sign-in/oauth/{provider}/finish"),
			Method:      http.MethodGet,
			HandlerFunc: i.finishOAuthSignIn,
		}})
	return nil
}

func (i Identity) verifyToken(w http.ResponseWriter, r *http.Request) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, isValid := i.identityService.VerifyAccessToken(string(buf))
	if isValid {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(strconv.Itoa(int(userID))))
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func (i Identity) oauthSignIn(w http.ResponseWriter, r *http.Request) {
	authProviderName := mux.Vars(r)["provider"]
	query := r.URL.Query()
	redirectURL := query.Get("redirectUrl")

	url, err := i.identityService.GenerateSignInURL(authProviderName, redirectURL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (i Identity) finishOAuthSignIn(w http.ResponseWriter, r *http.Request) {
	providerName := mux.Vars(r)["provider"]
	provider, err := i.identityService.GetOAuthProvider(providerName)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	stateID, err := provider.GetStateID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	authorizationCode := provider.GetAuthorizationCode(r)
	url, err := i.identityService.FinishOAuthSignIn(providerName, authorizationCode, stateID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func NewIdentity(identityService service.Identity) Identity {
	return Identity{
		identityService: identityService,
	}
}
