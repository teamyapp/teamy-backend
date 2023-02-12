package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type App interface {
	FindAppByAppID(ct context.Context, appID uint64) (entity.App, *errs.Error)
	FindApps(ct context.Context) ([]entity.App, *errs.Error)
	CreateApp(ct context.Context, app entity.App) *errs.Error
	UpdateApp(ct context.Context, app entity.App) *errs.Error
	DeleteApp(ct context.Context, appID uint64) *errs.Error
}
