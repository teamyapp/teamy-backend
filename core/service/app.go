package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
)

type App struct {
	dataCollector          telemetry.DataCollector
	cloudClientRegistry    *cloudAPI.ClientRegistry
	authorizer             Authorizer
	appTeamInstallationDao dao.AppTeamInstallation
}

type UpdateAppTeamInstallationInput struct {
	EnabledVersionNumber int32
}

func (a App) FindAppTeamInstallationsByAppId(ct context.Context, appID uint64) ([]entity.AppTeamInstallation, error) {
	return a.appTeamInstallationDao.FindAppTeamInstallationsByAppID(ct, appID)
}

func (a App) FindAppInstallationsByTeamId(ct context.Context, teamID uint64) ([]entity.AppTeamInstallation, error) {
	return a.appTeamInstallationDao.FindAppTeamInstallationsByTeamID(ct, teamID)
}

func (a App) CreateAppInstallation(ct context.Context, teamID uint64, appID uint64, versionNumber int32) (entity.AppTeamInstallation, error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppTeamInstallation{}, err
	}

	if feature.EnableAuthorization {
		query := authorization.NewCreateAppTeamInstallationQuery(userID, teamID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.authorizer.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppTeamInstallation{}, err
		}

		if !hasPermission {
			return entity.AppTeamInstallation{}, authorization.Error{
				Code:    authorization.UnauthorizedErrorCode,
				Message: fmt.Sprintf("Unauthorized: %v", query),
			}
		}
	}

	ai := entity.AppTeamInstallation{
		AppID:                appID,
		InstalledTeamID:      teamID,
		InstalledByUserID:    &userID,
		EnabledVersionNumber: versionNumber,
		InstalledAt:          time.Now(),
	}

	err := a.appTeamInstallationDao.CreateAppTeamInstallation(ct, ai)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppTeamInstallation{}, err
	}

	return ai, nil
}

func (a App) UpdateAppInstallation(ct context.Context, appID uint64, teamID uint64, input UpdateAppTeamInstallationInput) (entity.AppTeamInstallation, error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			err := errors.New("user id not found")
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppTeamInstallation{}, err
		}

		query := authorization.NewUpdateAppTeamInstallationQuery(userID, teamID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.authorizer.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppTeamInstallation{}, err
		}

		if !hasPermission {
			return entity.AppTeamInstallation{}, authorization.Error{
				Code:    authorization.UnauthorizedErrorCode,
				Message: fmt.Sprintf("Unauthorized: %v", query),
			}
		}
	}

	ai, err := a.appTeamInstallationDao.FindAppTeamInstallationByAppIDAndTeamID(ct, appID, teamID)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppTeamInstallation{}, err
	}

	ai.EnabledVersionNumber = input.EnabledVersionNumber
	err = a.appTeamInstallationDao.UpdateAppTeamInstallation(ct, ai)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppTeamInstallation{}, err
	}

	return ai, nil
}

func (a App) DeleteAppInstallation(ct context.Context, appID uint64, teamID uint64) (entity.AppTeamInstallation, error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			err := errors.New("user id not found")
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppTeamInstallation{}, err
		}

		query := authorization.NewDeleteAppTeamInstallationQuery(userID, teamID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.authorizer.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppTeamInstallation{}, err
		}

		if !hasPermission {
			return entity.AppTeamInstallation{}, authorization.Error{
				Code:    authorization.UnauthorizedErrorCode,
				Message: fmt.Sprintf("Unauthorized: %v", query),
			}
		}
	}

	ai, err := a.appTeamInstallationDao.FindAppTeamInstallationByAppIDAndTeamID(ct, appID, teamID)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppTeamInstallation{}, err
	}

	err = a.appTeamInstallationDao.DeleteAppTeamInstallation(ct, appID, teamID)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppTeamInstallation{}, err
	}

	return ai, nil
}

func NewApp(
	dataCollector telemetry.DataCollector,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	authorizer Authorizer,
	appTeamInstallationDao dao.AppTeamInstallation,
) App {
	return App{
		dataCollector,
		cloudClientRegistry,
		authorizer,
		appTeamInstallationDao,
	}
}
