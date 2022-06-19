package oauth

import (
	"net/http"

	"github.com/teamyapp/cloud/app/entity"
)

type Provider interface {
	GetName() string
	GetUser(authorizationCode string) (entity.ExternalUser, error)
	GetStateID(request *http.Request) (uint64, error)
	GetAuthorizationCode(request *http.Request) string
	GetSignInURL(stateID uint64) (string, error)
}
