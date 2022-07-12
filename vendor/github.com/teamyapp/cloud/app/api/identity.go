package api

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/web"
	"google.golang.org/grpc"
)

const identityPathPrefix = "/identity"

type Identity struct {
	identityService service.Identity
	proto.UnimplementedIdentityServer
}

var _ runner.Service = (*Identity)(nil)
var _ proto.IdentityServer = (*Identity)(nil)

func (i Identity) Start(rn *runner.ServiceRunner) error {
	rn.RegisterWebRoutes([]runner.WebRoute{
		{
			Path:        path.Join(identityPathPrefix, "verify-token"),
			Method:      http.MethodPost,
			HandlerFunc: i.webVerifyToken,
		},
		{
			Path:        path.Join(identityPathPrefix, "sign-in", "oauth", "{provider}"),
			Method:      http.MethodGet,
			HandlerFunc: i.webOauthSignIn,
		},
		{
			Path:        path.Join(identityPathPrefix, "sign-in", "oauth", "{provider}", "finish"),
			Method:      http.MethodGet,
			HandlerFunc: i.webFinishOAuthSignIn,
		},
		{
			Path:        path.Join(identityPathPrefix, "user-links"),
			Method:      http.MethodGet,
			HandlerFunc: i.webListUserLinks,
		},
		{
			Path:        path.Join(identityPathPrefix, "service-accounts"),
			Method:      http.MethodGet,
			HandlerFunc: i.webListServiceAccounts,
		},
		{
			Path:        path.Join(identityPathPrefix, "service-accounts", "create"),
			Method:      http.MethodPost,
			HandlerFunc: i.webCreateServiceAccount,
		},
		{
			Path:        path.Join(identityPathPrefix, "service-accounts", "{serviceAccountId}", "generate-token"),
			Method:      http.MethodPost,
			HandlerFunc: i.webGenerateServiceToken,
		},
		{
			Path:        path.Join(identityPathPrefix, "service-accounts", "{serviceAccountId}", "delete"),
			Method:      http.MethodDelete,
			HandlerFunc: i.webDeleteServiceAccount,
		},
	})
	rn.WithGRPCServer(func(server *grpc.Server) {
		proto.RegisterIdentityServer(server, i)
	})
	return nil
}

func (i Identity) webVerifyToken(w http.ResponseWriter, r *http.Request) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
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

func (i Identity) webOauthSignIn(w http.ResponseWriter, r *http.Request) {
	authProviderName := mux.Vars(r)["provider"]
	query := r.URL.Query()
	redirectURL := query.Get("redirectUrl")

	url, err := i.identityService.GenerateSignInURL(authProviderName, redirectURL)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (i Identity) webFinishOAuthSignIn(w http.ResponseWriter, r *http.Request) {
	providerName := mux.Vars(r)["provider"]
	provider, err := i.identityService.GetOAuthProvider(providerName)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	stateID, err := provider.GetStateID(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	authorizationCode := provider.GetAuthorizationCode(r)
	url, err := i.identityService.FinishOAuthSignIn(providerName, authorizationCode, stateID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (i Identity) webListUserLinks(writer http.ResponseWriter, request *http.Request) {
	userID, err := ctx.UserIDFromContext(request.Context())
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	userLinks, err := i.identityService.ListUserLinks(userID)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	web.WriteJSON(writer, userLinks)
}

func (i Identity) webListServiceAccounts(writer http.ResponseWriter, request *http.Request) {
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

	web.WriteJSON(writer, serviceAccounts)
}

func (i Identity) webCreateServiceAccount(writer http.ResponseWriter, request *http.Request) {
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

func (i Identity) webGenerateServiceToken(writer http.ResponseWriter, request *http.Request) {
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

func (i Identity) webDeleteServiceAccount(writer http.ResponseWriter, request *http.Request) {
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

func (i Identity) GetInternalUserId(ct context.Context, req *proto.GetInternalUserIdRequest) (*proto.GetInternalUserIdResponse, error) {
	internalUserID, err := i.identityService.GetInternalUserID(req.AuthProvider, req.ExternalUserId)
	if err != nil {
		return nil, err
	}

	return &proto.GetInternalUserIdResponse{InternalUserId: internalUserID}, nil
}

func NewIdentity(identityService service.Identity) Identity {
	return Identity{
		identityService: identityService,
	}
}
