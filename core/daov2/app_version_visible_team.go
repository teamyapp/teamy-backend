package daov2

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppVersionVisibleTeam interface {
	FindAppVersionVisibleTeamWithTx(
	    ct context.Context, 
	    tx *transaction.Transaction, 
	    appID uint64, versionNumber int32,
		teamID uint64,
	) (entity.AppVersionVisibleTeam, *errs.Error)
	FindAppVersionVisibleTeamsByAppIDAndVersionNumberWithTx(
	    ct context.Context, 
	    tx *transaction.Transaction,
		appID uint64, 
		versionNumber int32,
	) ([]entity.AppVersionVisibleTeam, *errs.Error)
	FindAppVersionVisibleTeamsByTeamIDWithTx(
	    ct context.Context, 
	    tx *transaction.Transaction,
		teamID uint64,
	) ([]entity.AppVersionVisibleTeam, *errs.Error)
	CreateAppVersionVisibleTeam(
	    ct context.Context, 
	    tx *transaction.Transaction,
		appVersionVisibleTeam entity.AppVersionVisibleTeam,
	) *errs.Error
	DeleteAppVersionVisibleTeam(
	    ct context.Context, 
	    tx *transaction.Transaction, 
	    appID uint64, 
	    versionNumber int32,
		teamID uint64,
	) *errs.Error
	DeleteAppVersionVisibleTeamsByAppIDAndVersionNumber(
	    ct context.Context, 
	    tx *transaction.Transaction, 
	    appID uint64,
		versionNumber int32,
	) *errs.Error
}
