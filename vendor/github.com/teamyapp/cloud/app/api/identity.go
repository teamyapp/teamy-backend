package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/web"
)

const identityPathPrefix = "/identity"

type Identity struct {
	identityService service.Identity
}

var _ runner.Service = (*Identity)(nil)

func (i Identity) Start(rn *runner.ServiceRunner) error {
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Path:        path.Join(identityPathPrefix, "verify-token"),
			Method:      http.MethodPost,
			HandlerFunc: i.verifyToken,
		},
		{
			Path:        path.Join(identityPathPrefix, "sign-in", "oauth", "{provider}"),
			Method:      http.MethodGet,
			HandlerFunc: i.oauthSignIn,
		},
		{
			Path:        path.Join(identityPathPrefix, "sign-in", "oauth", "{provider}", "finish"),
			Method:      http.MethodGet,
			HandlerFunc: i.finishOAuthSignIn,
		},
		{
			Path:        path.Join(identityPathPrefix, "service-accounts"),
			Method:      http.MethodGet,
			HandlerFunc: i.listServiceAccounts,
		},
		{
			Path:        path.Join(identityPathPrefix, "service-accounts", "create"),
			Method:      http.MethodPost,
			HandlerFunc: i.createServiceAccount,
		},
		{
			Path:        path.Join(identityPathPrefix, "service-accounts", "{serviceAccountId}", "generate-token"),
			Method:      http.MethodPost,
			HandlerFunc: i.generateServiceToken,
		},
		{
			Path:        path.Join(identityPathPrefix, "service-accounts", "{serviceAccountId}", "delete"),
			Method:      http.MethodDelete,
			HandlerFunc: i.deleteServiceAccount,
		},
	})
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

func (i Identity) listServiceAccounts(writer http.ResponseWriter, request *http.Request) {
	userID, err := ctx.UserIDFromContext(request.Context())
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	serviceAccounts, err := i.identityService.ListServiceAccounts(userID)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	buf, err := json.MarshalIndent(serviceAccounts, "", "  ")
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	web.WriteJSON(writer, buf)
}

func (i Identity) createServiceAccount(writer http.ResponseWriter, request *http.Request) {
	userID, err := ctx.UserIDFromContext(request.Context())
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	buf, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	err = json.Unmarshal(buf, &body)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = i.identityService.CreateServiceAccount(userID, body.Name)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func (i Identity) generateServiceToken(writer http.ResponseWriter, request *http.Request) {
	userID, err := ctx.UserIDFromContext(request.Context())
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	serviceAccountIDParam := mux.Vars(request)["serviceAccountId"]
	serviceAccountID, err := strconv.ParseUint(serviceAccountIDParam, 10, 64)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	serviceToken, err := i.identityService.GenerateServiceToken(userID, serviceAccountID)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.Write([]byte(serviceToken))
}

func (i Identity) deleteServiceAccount(writer http.ResponseWriter, request *http.Request) {
	userID, err := ctx.UserIDFromContext(request.Context())
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	serviceAccountIDParam := mux.Vars(request)["serviceAccountId"]
	serviceAccountID, err := strconv.ParseUint(serviceAccountIDParam, 10, 64)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	err = i.identityService.DeleteServiceAccount(userID, serviceAccountID)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

func NewIdentity(identityService service.Identity) Identity {
	return Identity{
		identityService: identityService,
	}
}
