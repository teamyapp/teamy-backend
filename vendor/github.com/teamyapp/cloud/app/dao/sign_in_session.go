package dao

import (
	"github.com/teamyapp/cloud/app/entity"
)

type SignInSession interface {
	FindSignInSessionByID(sessionID uint64) (entity.SignInSession, error)
	CreateSignInSession(session entity.SignInSession) error
	UpdateSignInSession(session entity.SignInSession) error
}
