package daov2

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppTeamInstallation interface {
	FindAppTeamInstallationsByAppID(ct context.Context, appID uint64) ([]entity.AppTeamInstallation, *errs.Error)
	FindAppTeamInstallationsByTeamID(ct context.Context, teamID uint64) ([]entity.AppTeamInstallation, *errs.Error)
	FindAppTeamInstallationsByAppIDWithTx(
		ct context.Context,
		tx *transaction.Transaction,
		appID uint64,
	) ([]entity.AppTeamInstallation, *errs.Error)
	FindAppTeamInstallationsByTeamIDWithTx(
		ct context.Context,
		tx *transaction.Transaction,
		teamID uint64,
	) ([]entity.AppTeamInstallation, *errs.Error)
	FindAppTeamInstallationByAppIDAndTeamIDWithTx(
		ct context.Context,
		tx *transaction.Transaction,
		appID uint64,
		teamID uint64,
	) (entity.AppTeamInstallation, *errs.Error)
	CreateAppTeamInstallation(
		ct context.Context,
		tx *transaction.Transaction,
		appTeamInstallation entity.AppTeamInstallation,
	) *errs.Error
	UpdateAppTeamInstallation(
		ct context.Context,
		tx *transaction.Transaction,
		appTeamInstallation entity.AppTeamInstallation,
	) *errs.Error
	DeleteAppTeamInstallation(
		ct context.Context,
		tx *transaction.Transaction,
		appID uint64,
		teamID uint64,
	) *errs.Error
	DeleteAppTeamInstallationsByAppIDAndVersionNumber(
		ct context.Context,
		tx *transaction.Transaction,
		appID uint64,
		versionNumber int32,
	) *errs.Error
}
