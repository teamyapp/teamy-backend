package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubAppInstallState interface {
	FindStateByID(ct context.Context, stateID uint64) (entity.GithubAppInstallState, *errs.Error)
	CreateState(ct context.Context, state entity.GithubAppInstallState) *errs.Error
	DeleteState(ct context.Context, stateID uint64) *errs.Error
}
