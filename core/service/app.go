package service

import (
	"context"
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
	IconURL                   *string
	HasUIExtension            bool
	UIExtensionEntryPointPath *string
	Changes                   *string
	IsPublic                  bool
}

type UpdateAppTeamInstallationInput struct {
	EnabledVersionNumber int32
}

func (a App) FindAppTeamInstallationsByAppId(ct context.Context, appID uint64) ([]entity.AppTeamInstallation, *errs.Error) {
	return a.appTeamInstallationDao.FindAppTeamInstallationsByAppID(ct, appID)
}

func (a App) FindAppInstallationsByTeamId(ct context.Context, teamID uint64) ([]entity.AppTeamInstallation, *errs.Error) {
	return a.appTeamInstallationDao.FindAppTeamInstallationsByTeamID(ct, teamID)
}

func (a App) FindAppVersionByAppId(ct context.Context, appID uint64) ([]entity.AppVersion, *errs.Error) {
	return a.appVersionDao.FindAppVersionsByAppID(ct, appID)
}

func (a App) FindAppVersionByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int32) (entity.AppVersion, *errs.Error) {
	return a.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
}

func (a App) FindAppVersionVisibleTeams(ct context.Context, appID uint64, versionNumber int32) ([]entity.AppVersionVisibleTeam, *errs.Error) {
	return a.appVersionVisibleTeamDao.FindAppVersionVisibleTeamsByAppIDAndVersionNumber(ct, appID, versionNumber)
}

func (a App) CreateAppVersion(ct context.Context, appID uint64) (entity.AppVersion, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: "user ID not found",
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.AppVersion{}, internalErr
	}

	if feature.EnableAuthorization {
		query := authorization.NewCreateAppVersionQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.authorizer.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("authorization query: %v", query),
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return entity.AppVersion{}, internalErr
		}
	}

	av := entity.AppVersion{
		AppID:          appID,
		HasUIExtension: false,
		IsPublic:       false,
		CreatedAt:      time.Now().UTC(),
	}
	versionNum, err := a.appVersionDao.FindMaxVersionNumber(ct, appID)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppVersion{}, err
	}
	av.VersionNumber = versionNum + 1

	err = a.appVersionDao.CreateAppVersion(ct, av)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppVersion{}, err
	}

	return av, nil
}

func (a App) UpdateAppVersion(ct context.Context, appID uint64, versionNumber int32, input UpdateAppVersionInput) (entity.AppVersion, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			internalErr := &errs.Error{
				Code:    errs.NotFound,
				Message: "user ID not found",
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return entity.AppVersion{}, internalErr
		}

		query := authorization.NewUpdateAppVersionQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.authorizer.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("authorization query: %v", query),
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return entity.AppVersion{}, internalErr
		}
	}

	av, err := a.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppVersion{}, err
	}

	av.HasUIExtension = input.HasUIExtension
	av.IsPublic = input.IsPublic
	if input.IconURL != nil {
		av.IconURL = input.IconURL
	}
	if input.Changes != nil {
		av.Changes = input.Changes
	}
	if input.UIExtensionEntryPointPath != nil {
		av.UIExtensionEntrypointPath = input.UIExtensionEntryPointPath
	}
	now := time.Now().UTC()
	av.UpdateAt = &now
	err = a.appVersionDao.UpdateAppVersion(ct, av)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppVersion{}, err
	}

	return av, nil
}

func (a App) DeleteAppVersion(ct context.Context, appID uint64, versionNumber int32) (entity.AppVersion, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			internalErr := &errs.Error{
				Code:    errs.NotFound,
				Message: "user ID not found",
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return entity.AppVersion{}, internalErr
		}

		query := authorization.NewDeleteAppVersionQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.authorizer.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("authorization query: %v", query),
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return entity.AppVersion{}, internalErr
		}
	}

	av, err := a.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppVersion{}, err
	}

	// TODO(yuhang): check if version to delete is the active version of the app after implementing app APIs

	// TODO(yumiao): Below operations should be atomic by wrapping in one txn. We'll implement that after adding our
	// own transaction library.
	err = a.appVersionDao.DeleteAppVersion(ct, appID, versionNumber)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	err = a.appTeamInstallationDao.DeleteAppTeamInstallationByAppIDAndVersionNumber(ct, appID, versionNumber)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppVersion{}, err
	}

	err = a.appVersionVisibleTeamDao.DeleteAppVersionVisibleTeamByAppIDAndVersionNumber(ct, appID, versionNumber)
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
			internalErr := &errs.Error{
				Code:    errs.NotFound,
				Message: "user ID not found",
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return entity.AppVersionVisibleTeam{}, internalErr
		}

		query := authorization.NewCreateAppVersionVisibleTeamQuery(userID, appID)
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
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppVersionVisibleTeam{}, err
	}

	return av, nil
}

func (a App) DeleteAppVersionVisibleTeam(ct context.Context, appID uint64, versionNumber int32, teamID uint64) (entity.AppVersionVisibleTeam, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			internalErr := &errs.Error{
				Code:    errs.NotFound,
				Message: "user ID not found",
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return entity.AppVersionVisibleTeam{}, internalErr
		}

		query := authorization.NewDeleteAppVersionVisibleTeamQuery(userID, appID)
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
