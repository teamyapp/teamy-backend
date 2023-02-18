package service

import (
	"context"
	"fmt"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/collect"
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
	appDao                   dao.App
	appVersionDao            dao.AppVersion
	appTeamInstallationDao   dao.AppTeamInstallation
	appVersionVisibleTeamDao dao.AppVersionVisibleTeam
	teamDao                  dao.Team
}

type AppFilter struct {
	AppID  *uint64
	TeamID *uint64
}

type UpdateAppInput struct {
	AppName             *string
	Description         *string
	ActiveVersionNumber *int32
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

func (a App) FindAppByID(ct context.Context, appID uint64) (entity.App, *errs.Error) {
	return a.appDao.FindAppByID(ct, appID)
}

func (a App) FindApps(ct context.Context, filter *AppFilter) ([]entity.App, *errs.Error) {
	var apps []entity.App
	if filter != nil {
		if filter.AppID != nil {
			app, err := a.appDao.FindAppByID(ct, *filter.AppID)
			if err != nil {
				a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
				return nil, err
			}
			apps = append(apps, app)
		} else {
			var err *errs.Error
			apps, err = a.appDao.FindAllApps(ct)
			if err != nil {
				a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
				return nil, err
			}
		}

		if filter.TeamID != nil {
			// find all apps that are visible to the team
			apps = collect.Filter(apps, func(app entity.App) bool {
				appVersions, err := a.appVersionDao.FindAppVersionsByAppID(ct, app.ID)
				if err != nil {
					a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
					return false
				}

				appVersions = collect.Filter(appVersions, func(appVersion entity.AppVersion) bool {
					if app.ActiveVersionNumber != nil && appVersion.VersionNumber < *app.ActiveVersionNumber {
						// if active version has been set, we should filter all old versions
						return false
					}

					if !appVersion.IsPublic {
						_, err := a.appVersionVisibleTeamDao.FindAppVersionVisibleTeam(ct, app.ID, appVersion.VersionNumber, *filter.TeamID)
						if err != nil {
							if err.Code != errs.NotFound {
								a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
							}
							return false
						}
					}

					return true
				})

				if len(appVersions) > 0 {
					return true
				}

				return false
			})
		}
	} else {
		var err *errs.Error
		apps, err = a.appDao.FindAllApps(ct)
		if err != nil {
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return nil, err
		}
	}

	return apps, nil
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

func (a App) FindAppVersionVisibleTeams(ct context.Context, appID uint64, versionNumber int32) ([]entity.Team, *errs.Error) {
	appVersionVisibleTeams, err := a.appVersionVisibleTeamDao.FindAppVersionVisibleTeamsByAppIDAndVersionNumber(ct, appID, versionNumber)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, err
	}

	teamIDs := collect.Map(appVersionVisibleTeams, func(appVersionVisibleTeam entity.AppVersionVisibleTeam, _ int) uint64 {
		return appVersionVisibleTeam.TeamID
	})
	teams, err := a.teamDao.FindTeamsByIDs(ct, teamIDs)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, err
	}

	return teams, nil
}

func (a App) CreateApp(ct context.Context, name string) (entity.App, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.App{}, internalErr
	}

	genAppIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "appID"}
	genAppIDRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genAppIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.App{}, internalErr
	}

	genAppSecretReq := &proto.GenerateUniqueStringRequest{SequenceName: "apiSecret"}
	genAppSecretRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueString(ct, genAppSecretReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.App{}, internalErr
	}

	app := entity.App{
		ID:                genAppIDRes.UniqueNumber,
		Name:              name,
		Description:       "",
		APISecret:         genAppSecretRes.UniqueString,
		InstallationCount: 0,
		CreatorUserID:     userID,
		CreatedAt:         time.Now(),
	}

	err := a.appDao.CreateApp(ct, app)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.App{}, err
	}

	if feature.EnableAuthorization {
		err = a.authorizer.registerResource(ct, authorization.AppResourceType, app.ID)
		if err != nil {
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.App{}, err
		}

		// When create a new app,
		// 1) Resource: register app resource
		// 2) UserGroup: create AppAdmin and AppMember userGroups
		// 3) UserGroupMember:
		// 		add app creator to AppAdmin group
		// 		add app creator to AppMember group
		// 4) Permissions:
		//		assign AppAdmin permissions to app creator
		//		assign AppMember permissions to app creator
		appAdminUserGroupName := fmt.Sprintf("App%d/Admin", app.ID)
		appAdminDescription := fmt.Sprintf("Admins for %s", appAdminUserGroupName)
		appAdminOperations := make([]authorization.ResourceOperation, 0)
		for _, appAdminResourceTypeOperation := range authorization.AppAdminResourceTypeOperations {
			appAdminOperations = append(appAdminOperations, authorization.ResourceOperation{
				ResourceType: appAdminResourceTypeOperation.ResourceType,
				Operation:    appAdminResourceTypeOperation.Operation,
				ResourceID:   app.ID,
			})
		}

		_, err := a.authorizer.createUserGroupAndAssignPermissions(ct,
			userID,
			appAdminUserGroupName,
			&appAdminDescription,
			appAdminOperations,
		)
		if err != nil {
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.App{}, err
		}

		appMemberUserGroupName := fmt.Sprintf("App%d/Member", app.ID)
		appMemberDescription := fmt.Sprintf("Members for %s", appMemberUserGroupName)
		appMemberOperations := make([]authorization.ResourceOperation, 0)
		for _, appMemberResourceTypeOperation := range authorization.AppMemberResourceTypeOperations {
			appMemberOperations = append(appMemberOperations, authorization.ResourceOperation{
				ResourceType: appMemberResourceTypeOperation.ResourceType,
				Operation:    appMemberResourceTypeOperation.Operation,
				ResourceID:   app.ID,
			})
		}

		_, err = a.authorizer.createUserGroupAndAssignPermissions(ct,
			userID,
			appMemberUserGroupName,
			&appMemberDescription,
			appMemberOperations,
		)
		if err != nil {
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.App{}, err
		}
	}

	return app, nil
}

func (a App) UpdateApp(ct context.Context, appID uint64, input UpdateAppInput) (entity.App, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			internalErr := &errs.Error{
				Code:    errs.Unauthenticated,
				Message: "user ID not found",
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return entity.App{}, internalErr
		}

		query := authorization.NewUpdateAppQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.authorizer.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.App{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("authorization query: %v", query),
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return entity.App{}, internalErr
		}
	}

	app, err := a.appDao.FindAppByID(ct, appID)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.App{}, err
	}

	if input.AppName != nil {
		app.Name = *input.AppName
	}
	if input.Description != nil {
		app.Description = *input.Description
	}
	if input.ActiveVersionNumber != nil {
		if app.ActiveVersionNumber != nil {
			if *app.ActiveVersionNumber > *input.ActiveVersionNumber {
				internalErr := &errs.Error{
					Code: errs.InvalidOperation,
					Message: fmt.Sprintf(
						"Roll back active version is invalid: stateID=%v", appID),
				}
				a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
				return entity.App{}, internalErr
			}

			if *app.ActiveVersionNumber < *input.ActiveVersionNumber {
				// roll forward app installation automatically
				a.RollForwardAppInstallations(ct, appID, *input.ActiveVersionNumber)
			}
		}

		app.ActiveVersionNumber = input.ActiveVersionNumber
	}

	now := time.Now().UTC()
	app.UpdatedAt = &now
	err = a.appDao.UpdateApp(ct, app)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.App{}, err
	}

	return app, nil
}

func (a App) RefreshAppSecret(ct context.Context, appID uint64) (entity.App, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			internalErr := &errs.Error{
				Code:    errs.Unauthenticated,
				Message: "user ID not found",
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return entity.App{}, internalErr
		}

		query := authorization.NewRefreshAppSecretQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.authorizer.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.App{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("authorization query: %v", query),
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return entity.App{}, internalErr
		}
	}

	app, err := a.appDao.FindAppByID(ct, appID)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.App{}, err
	}

	genAppSecretReq := &proto.GenerateUniqueStringRequest{SequenceName: "apiSecret"}
	genAppSecretRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueString(ct, genAppSecretReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.App{}, internalErr
	}

	app.APISecret = genAppSecretRes.UniqueString
	err = a.appDao.UpdateApp(ct, app)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.App{}, err
	}

	return app, nil
}

func (a App) DeleteApp(ct context.Context, appID uint64) (entity.App, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			internalErr := &errs.Error{
				Code:    errs.Unauthenticated,
				Message: "user ID not found",
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return entity.App{}, internalErr
		}

		query := authorization.NewDeleteAppQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.authorizer.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return entity.App{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("authorization query: %v", query),
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			return entity.App{}, internalErr
		}
	}

	app, err := a.appDao.FindAppByID(ct, appID)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.App{}, err
	}

	err = a.appDao.DeleteApp(ct, appID)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.App{}, err
	}

	// TODO(yuhang): delete registered app resource and groups

	return app, nil
}

func (a App) CreateAppVersion(ct context.Context, appID uint64) (entity.AppVersion, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.AppVersion{}, internalErr
	}

	if feature.EnableAuthorization {
		query := authorization.NewCreateAppVersionQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.authorizer.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("permission denied: authorization query=%v", query),
			}
			a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.AppVersion{}, internalErr
		}
	}

	av := entity.AppVersion{
		AppID:          appID,
		HasUIExtension: false,
		IsPublic:       false,
		CreatedAt:      time.Now().UTC(),
	}
	maxVersion, err := a.appVersionDao.FindMaxVersionNumber(ct, appID)
	if err != nil {
		if err.Code == errs.NotFound {
			// no version exists, start from 0
			maxVersion = 0
		} else {
			a.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.AppVersion{}, err
		}
	}

	av.VersionNumber = maxVersion + 1
	err = a.appVersionDao.CreateAppVersion(ct, av)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.AppVersion{}, err
	}

	return av, nil
}

func (a App) UpdateAppVersion(ct context.Context, appID uint64, versionNumber int32, input UpdateAppVersionInput) (entity.AppVersion, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			internalErr := &errs.Error{
				Code:    errs.Unauthenticated,
				Message: "user ID not found",
			}
			a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.AppVersion{}, internalErr
		}

		query := authorization.NewUpdateAppVersionQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.authorizer.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("permission denied: authorization query=%v", query),
			}
			a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.AppVersion{}, internalErr
		}
	}

	av, err := a.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.AppVersion{}, err
	}

	av.HasUIExtension = input.HasUIExtension
	av.IsPublic = input.IsPublic
	av.IconURL = input.IconURL
	av.Changes = input.Changes
	av.UIExtensionEntrypointPath = input.UIExtensionEntryPointPath
	now := time.Now().UTC()
	av.UpdateAt = &now
	err = a.appVersionDao.UpdateAppVersion(ct, av)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.AppVersion{}, err
	}

	return av, nil
}

func (a App) DeleteAppVersion(ct context.Context, appID uint64, versionNumber int32) (entity.AppVersion, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			internalErr := &errs.Error{
				Code:    errs.Unauthenticated,
				Message: "user ID not found",
			}
			a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.AppVersion{}, internalErr
		}

		query := authorization.NewDeleteAppVersionQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.authorizer.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("permission denied: authorization query=%v", query),
			}
			a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.AppVersion{}, internalErr
		}
	}

	av, err := a.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.AppVersion{}, err
	}

	app, err := a.appDao.FindAppByID(ct, appID)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.AppVersion{}, err
	}

	if app.ActiveVersionNumber != nil && *app.ActiveVersionNumber == versionNumber {
		internalErr := &errs.Error{
			Code: errs.InvalidOperation,
			Message: fmt.Sprintf(
				"Cannot delete active version: appID=%v", appID),
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.AppVersion{}, internalErr
	}

	// TODO(yuhang): below operations should be atomic by wrapping in one txn. We'll implement that after adding our
	// own transaction library.
	err = a.appVersionDao.DeleteAppVersion(ct, appID, versionNumber)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
	}

	err = a.appTeamInstallationDao.DeleteAppTeamInstallationsByAppIDAndVersionNumber(ct, appID, versionNumber)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.AppVersion{}, err
	}

	err = a.appVersionVisibleTeamDao.DeleteAppVersionVisibleTeamsByAppIDAndVersionNumber(ct, appID, versionNumber)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.AppVersion{}, err
	}

	return av, nil
}

func (a App) CreateAppVersionVisibleTeam(ct context.Context, appID uint64, versionNumber int32, teamID uint64) (entity.AppVersion, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			internalErr := &errs.Error{
				Code:    errs.Unauthenticated,
				Message: "user ID not found",
			}
			a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.AppVersion{}, internalErr
		}

		query := authorization.NewCreateAppVersionVisibleTeamQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.authorizer.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("permission denied: authorization query=%v", query),
			}
			a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.AppVersion{}, internalErr
		}
	}

	av := entity.AppVersionVisibleTeam{
		AppID:         appID,
		VersionNumber: versionNumber,
		TeamID:        teamID,
	}
	err := a.appVersionVisibleTeamDao.CreateAppVersionVisibleTeam(ct, av)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.AppVersion{}, err
	}

	appVersion, err := a.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.AppVersion{}, err
	}

	return appVersion, nil
}

func (a App) DeleteAppVersionVisibleTeam(ct context.Context, appID uint64, versionNumber int32, teamID uint64) (entity.AppVersion, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			internalErr := &errs.Error{
				Code:    errs.Unauthenticated,
				Message: "user ID not found",
			}
			a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.AppVersion{}, internalErr
		}

		query := authorization.NewDeleteAppVersionVisibleTeamQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.authorizer.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("permission denied: authorization query=%v", query),
			}
			a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.AppVersion{}, internalErr
		}
	}

	av, err := a.appVersionVisibleTeamDao.FindAppVersionVisibleTeam(ct, appID, versionNumber, teamID)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.AppVersion{}, err
	}

	err = a.appVersionVisibleTeamDao.DeleteAppVersionVisibleTeam(ct, appID, versionNumber, teamID)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.AppVersion{}, err
	}

	// if team has installed the version, delete installation as well
	appTeamInstallation, err := a.appTeamInstallationDao.FindAppTeamInstallationByAppIDAndTeamID(ct, appID, teamID)
	if err != nil {
		if err.Code != errs.NotFound {
			a.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.AppVersion{}, err
		}
	} else {
		if appTeamInstallation.EnabledVersionNumber == versionNumber {
			err = a.appTeamInstallationDao.DeleteAppTeamInstallation(ct, appID, teamID)
			if err != nil {
				a.dataCollector.Logger.ErrorWithContext(ct, err)
				return entity.AppVersion{}, err
			}
		}
	}

	appVersion, err := a.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, av.AppID, av.VersionNumber)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.AppVersion{}, err
	}

	return appVersion, nil
}

func (a App) CreateAppInstallation(ct context.Context, teamID uint64, appID uint64, versionNumber int32) (entity.AppTeamInstallation, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.AppTeamInstallation{}, internalErr
	}

	if feature.EnableAuthorization {
		query := authorization.NewCreateAppTeamInstallationQuery(userID, teamID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.AppTeamInstallation{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("permission denied: authorization query=%v", query),
			}
			a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.AppTeamInstallation{}, internalErr
		}
	}

	app, err := a.appDao.FindAppByID(ct, appID)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.AppTeamInstallation{}, err
	}

	ai := entity.AppTeamInstallation{
		AppID:                appID,
		InstalledTeamID:      teamID,
		InstalledByUserID:    &userID,
		EnabledVersionNumber: versionNumber,
		InstalledAt:          time.Now(),
	}

	err = a.appTeamInstallationDao.CreateAppTeamInstallation(ct, ai)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.AppTeamInstallation{}, err
	}

	app.InstallationCount = app.InstallationCount + 1
	err = a.appDao.UpdateApp(ct, app)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.AppTeamInstallation{}, err
	}

	return ai, nil
}

func (a App) UpdateAppInstallation(ct context.Context, appID uint64, teamID uint64, input UpdateAppTeamInstallationInput) (entity.AppTeamInstallation, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			internalErr := &errs.Error{
				Code:    errs.Unauthenticated,
				Message: "user ID not found",
			}
			a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.AppTeamInstallation{}, internalErr
		}

		query := authorization.NewUpdateAppTeamInstallationQuery(userID, teamID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.AppTeamInstallation{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("permission denied: authorization query=%v", query),
			}
			a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.AppTeamInstallation{}, internalErr
		}
	}

	ai, err := a.appTeamInstallationDao.FindAppTeamInstallationByAppIDAndTeamID(ct, appID, teamID)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.AppTeamInstallation{}, err
	}

	ai.EnabledVersionNumber = input.EnabledVersionNumber
	err = a.appTeamInstallationDao.UpdateAppTeamInstallation(ct, ai)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.AppTeamInstallation{}, err
	}

	return ai, nil
}

func (a App) DeleteAppInstallation(ct context.Context, appID uint64, teamID uint64) (entity.AppTeamInstallation, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			internalErr := &errs.Error{
				Code:    errs.Unauthenticated,
				Message: "user ID not found",
			}
			a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.AppTeamInstallation{}, internalErr
		}

		query := authorization.NewDeleteAppTeamInstallationQuery(userID, teamID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			a.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.AppTeamInstallation{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("permission denied: authorization query=%v", query),
			}
			a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.AppTeamInstallation{}, internalErr
		}
	}

	ai, err := a.appTeamInstallationDao.FindAppTeamInstallationByAppIDAndTeamID(ct, appID, teamID)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.AppTeamInstallation{}, err
	}

	err = a.appTeamInstallationDao.DeleteAppTeamInstallation(ct, appID, teamID)
	if err != nil {
		a.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.AppTeamInstallation{}, err
	}

	return ai, nil
}

// RollForwardAppInstallations Helper function to roll forward all app installations when a
// newer active version enabled
func (a App) RollForwardAppInstallations(ct context.Context, appID uint64, activeVersionNumber int32) *errs.Error {
	appInstallations, err := a.appTeamInstallationDao.FindAppTeamInstallationsByAppID(ct, appID)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return err
	}

	for _, appInstallation := range appInstallations {
		if appInstallation.EnabledVersionNumber < activeVersionNumber {
			appInstallation.EnabledVersionNumber = activeVersionNumber
			err = a.appTeamInstallationDao.UpdateAppTeamInstallation(ct, appInstallation)
			if err != nil {
				a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			}
		}
	}

	return nil
}

func NewApp(
	dataCollector telemetry.DataCollector,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	authorizer Authorizer,
	appDao dao.App,
	appVersionDao dao.AppVersion,
	appTeamInstallationDao dao.AppTeamInstallation,
	appVersionVisibleTeamDao dao.AppVersionVisibleTeam,
	teamDao dao.Team,
) App {
	return App{
		dataCollector,
		cloudClientRegistry,
		authorizer,
		appDao,
		appVersionDao,
		appTeamInstallationDao,
		appVersionVisibleTeamDao,
		teamDao,
	}
}
