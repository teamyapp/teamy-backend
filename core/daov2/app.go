package daov2

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type App interface {
	FindAppByID(ct context.Context, appID uint64) (entity.App, *errs.Error)
	FindAllApps(ct context.Context) ([]entity.App, *errs.Error)
	FindAppByIDWithTx(ct context.Context, tx *transaction.Transaction, appID uint64) (entity.App, *errs.Error)
	FindAllAppsWithTx(ct context.Context, tx *transaction.Transaction) ([]entity.App, *errs.Error)
	CreateApp(ct context.Context, tx *transaction.Transaction, app entity.App) *errs.Error
	UpdateApp(ct context.Context, tx *transaction.Transaction, app entity.App) *errs.Error
	DeleteApp(ct context.Context, tx *transaction.Transaction, appID uint64) *errs.Error
}
