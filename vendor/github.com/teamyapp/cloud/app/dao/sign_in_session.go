package dao

import (
	"github.com/teamyapp/cloud/app/entity"
)

type SignInSession interface {
	FindByID(sessionID uint64) (entity.SignInSession, error)
	Add(session entity.SignInSession) error
	Update(session entity.SignInSession) error
}
