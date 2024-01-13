package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamAppInstallation interface {
	FindTeamAppInstallationsByAppIDWithTx(ct context.Context, tx *transaction.Transaction, appID uint64) ([]entity.TeamAppInstallation, *errs.Error)
	FindTeamAppInstallationsByAppID(ct context.Context, appID uint64) ([]entity.TeamAppInstallation, *errs.Error)
	FindTeamAppInstallationsByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.TeamAppInstallation, *errs.Error)
	FindTeamAppInstallationsByTeamID(ct context.Context, teamID uint64) ([]entity.TeamAppInstallation, *errs.Error)
	FindTeamAppInstallationByIDWithTx(ct context.Context, tx *transaction.Transaction, appInstallationID uint64) (entity.TeamAppInstallation, *errs.Error)
	CreateTeamAppInstallation(ct context.Context, tx *transaction.Transaction, teamAppInstallation entity.TeamAppInstallation) *errs.Error
	DeleteTeamAppInstallationByID(ct context.Context, tx *transaction.Transaction, appInstallationID uint64) *errs.Error
}
