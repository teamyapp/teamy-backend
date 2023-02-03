package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
)

type App struct {
	dataCollector            telemetry.DataCollector
	cloudClientRegistry      *cloudAPI.ClientRegistry
	authorizer               Authorizer
	appVersionDao            dao.AppVersion
	appTeamInstallationDao   dao.AppTeamInstallation
	appVersionVisibleTeamDao dao.AppVersionVisibleTeam
}

type UpdateAppVersionInput struct {
	IconUrl                   *string
	HasUiExtension            bool
	UiExtensionEntryPointPath *string
	Changes                   *string
	IsPublic                  bool
}

type UpdateAppTeamInstallationInput struct {
	EnabledVersionNumber int32
}

func (a App) FindAppVersionByAppId(ct context.Context, appID uint64) ([]entity.AppVersion, error) {
	return a.appVersionDao.FindAppVersionsByAppID(ct, appID)
}

func (a App) FindAppVersionByAppIdAndVersionNumber(ct context.Context, appID uint64, versionNumber int32) (entity.AppVersion, error) {
	return a.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
}

func (a App) FindAppVersionVisibleTeams(ct context.Context, appID uint64, versionNumber int32) ([]entity.AppVersionVisibleTeam, error) {
	return a.appVersionVisibleTeamDao.FindAppVersionVisibleTeamsByAppIDAndVersionNumber(ct, appID, versionNumber)
}

func (a App) CreateAppVersion(ct context.Context, appID uint64) (entity.AppVersion, error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppVersion{}, err
	}

	if feature.EnableAuthorization {
		query := authorization.NewCreateAppVersionQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.authorizer.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			return entity.AppVersion{}, authorization.Error{
				Code:    authorization.UnauthorizedErrorCode,
				Message: fmt.Sprintf("Unauthorized: %v", query),
			}
		}
	}

	av := entity.AppVersion{
		AppID:          appID,
		HasUiExtension: false,
		IsPublic:       false,
		CreatedAt:      time.Now(),
	}

	versionNumber, err := a.appVersionDao.CreateAppVersion(ct, av)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppVersion{}, err
	}
	av.VersionNumber = versionNumber

	return av, nil
}

func (a App) UpdateAppVersion(ct context.Context, appID uint64, versionNumber int32, input UpdateAppVersionInput) (entity.AppVersion, error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			err := errors.New("user id not found")
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppVersion{}, err
		}

		query := authorization.NewUpdateAppTeamInstallationQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.authorizer.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			return entity.AppVersion{}, authorization.Error{
				Code:    authorization.UnauthorizedErrorCode,
				Message: fmt.Sprintf("Unauthorized: %v", query),
			}
		}
	}

	av, err := a.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppVersion{}, err
	}

	av.HasUiExtension = input.HasUiExtension
	av.IsPublic = input.IsPublic
	if input.IconUrl != nil {
		av.IconUrl = input.IconUrl
	}
	if input.Changes != nil {
		av.Changes = input.Changes
	}
	if input.UiExtensionEntryPointPath != nil {
		av.UiExtensionEntrypointPath = input.UiExtensionEntryPointPath
	}
	now := time.Now()
	av.UpdateAt = &now
	err = a.appVersionDao.UpdateAppVersion(ct, av)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppVersion{}, err
	}

	return av, nil
}

func (a App) DeleteAppVersion(ct context.Context, appID uint64, versionNumber int32) (entity.AppVersion, error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			err := errors.New("user id not found")
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppVersion{}, err
		}

		query := authorization.NewDeleteAppTeamInstallationQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.authorizer.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			return entity.AppVersion{}, authorization.Error{
				Code:    authorization.UnauthorizedErrorCode,
				Message: fmt.Sprintf("Unauthorized: %v", query),
			}
		}
	}

	av, err := a.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppVersion{}, err
	}

	// TODO(yuhang): check if version to delete is the active version of the app after implementing app APIs

	err = a.appVersionDao.DeleteAppVersion(ct, appID, versionNumber)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppVersion{}, err
	}

	return av, nil
}

func (a App) CreateAppVersionVisibleTeam(ct context.Context, appID uint64, versionNumber int32, teamID uint64) (entity.AppVersionVisibleTeam, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			err := errors.New("user id not found")
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppVersionVisibleTeam{}, err
		}

		query := authorization.NewDeleteAppTeamInstallationQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.authorizer.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppVersionVisibleTeam{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("authorization query: %v", query),
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return entity.AppVersionVisibleTeam{}, internalErr
		}
	}

	av := entity.AppVersionVisibleTeam{
		AppID:         appID,
		VersionNumber: versionNumber,
		TeamID:        teamID,
	}

	err := a.appVersionVisibleTeamDao.CreateAppVersionVisibleTeam(ct, av)
	if err != nil {
		a.authorizer.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppVersionVisibleTeam{}, err
	}

	return av, nil
}

func (a App) DeleteAppVersionVisibleTeam(ct context.Context, appID uint64, versionNumber int32, teamID uint64) (entity.AppVersionVisibleTeam, error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			err := errors.New("user id not found")
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppVersionVisibleTeam{}, err
		}

		query := authorization.NewDeleteAppTeamInstallationQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.authorizer.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppVersionVisibleTeam{}, err
		}

		if !hasPermission {
			return entity.AppVersionVisibleTeam{}, authorization.Error{
				Code:    authorization.UnauthorizedErrorCode,
				Message: fmt.Sprintf("Unauthorized: %v", query),
			}
		}
	}

	av, err := a.appVersionVisibleTeamDao.FindAppVersionVisibleTeam(ct, appID, versionNumber, teamID)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppVersionVisibleTeam{}, err
	}

	err = a.appVersionVisibleTeamDao.DeleteAppVersionVisibleTeam(ct, appID, versionNumber, teamID)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppVersionVisibleTeam{}, err
	}

	return av, nil
}

func (a App) CreateAppInstallation(ct context.Context, teamID uint64, appID uint64, versionNumber int32) (entity.AppTeamInstallation, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: "user ID not found",
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.AppTeamInstallation{}, internalErr
	}

	if feature.EnableAuthorization {
		query := authorization.NewCreateAppTeamInstallationQuery(userID, teamID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppTeamInstallation{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("authorization query: %v", query),
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return entity.AppTeamInstallation{}, internalErr
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

func (a App) UpdateAppInstallation(ct context.Context, appID uint64, teamID uint64, input UpdateAppTeamInstallationInput) (entity.AppTeamInstallation, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			internalErr := &errs.Error{
				Code:    errs.NotFound,
				Message: "user ID not found",
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return entity.AppTeamInstallation{}, internalErr
		}

		query := authorization.NewUpdateAppTeamInstallationQuery(userID, teamID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppTeamInstallation{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("authorization query: %v", query),
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return entity.AppTeamInstallation{}, internalErr
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

func (a App) DeleteAppInstallation(ct context.Context, appID uint64, teamID uint64) (entity.AppTeamInstallation, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			internalErr := &errs.Error{
				Code:    errs.NotFound,
				Message: "user ID not found",
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return entity.AppTeamInstallation{}, internalErr
		}

		query := authorization.NewDeleteAppTeamInstallationQuery(userID, teamID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppTeamInstallation{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("authorization query: %v", query),
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return entity.AppTeamInstallation{}, internalErr
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
	appVersionDao dao.AppVersion,
	appTeamInstallationDao dao.AppTeamInstallation,
	appVersionVisibleTeamDao dao.AppVersionVisibleTeam,
) App {
	return App{
		dataCollector,
		cloudClientRegistry,
		authorizer,
		appVersionDao,
		appTeamInstallationDao,
		appVersionVisibleTeamDao,
	}
}
