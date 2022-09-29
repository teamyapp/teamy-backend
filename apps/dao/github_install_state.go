package dao

import (
	"context"

	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubAppInstallState interface {
	FindStateByID(ct context.Context, stateID uint64) (entity.GithubAppInstallState, error)
	CreateState(ct context.Context, state entity.GithubAppInstallState) error
	DeleteState(ct context.Context, stateID uint64) error
}
