package dao

import (
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubAppInstallState interface {
	FindStateByID(stateID uint64) (entity.GithubAppInstallState, error)
	CreateState(state entity.GithubAppInstallState) error
	DeleteState(stateID uint64) error
}
