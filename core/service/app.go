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
	logger                   telemetry.Logger
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
	Name                *string
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
				return nil, err
			}

			apps = append(apps, app)
		} else {
			var err *errs.Error
			apps, err = a.appDao.FindAllApps(ct)
			if err != nil {
				return nil, err
			}
		}

		var err *errs.Error
		if filter.TeamID != nil {
			// find all apps that are visible to the team
			apps = collect.Filter(apps, func(app entity.App) bool {
				var visible bool
				visible, err = a.isAppVisibleToTeam(ct, app, *filter.TeamID)
				return visible
			})
		}

	} else {
		var err *errs.Error
		apps, err = a.appDao.FindAllApps(ct)
		if err != nil {
			return nil, err
		}
	}

	return apps, nil
}

func (a App) FindAppTeamInstallationsByAppID(ct context.Context, appID uint64) ([]entity.AppTeamInstallation, *errs.Error) {
	return a.appTeamInstallationDao.FindAppTeamInstallationsByAppID(ct, appID)
}

func (a App) FindAppInstallationsByTeamID(ct context.Context, teamID uint64) ([]entity.AppTeamInstallation, *errs.Error) {
	return a.appTeamInstallationDao.FindAppTeamInstallationsByTeamID(ct, teamID)
}

func (a App) FindAppVersionByAppID(ct context.Context, appID uint64) ([]entity.AppVersion, *errs.Error) {
	return a.appVersionDao.FindAppVersionsByAppID(ct, appID)
}

func (a App) FindAppVersionByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int32) (entity.AppVersion, *errs.Error) {
	return a.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
}

func (a App) FindAppVersionVisibleTeams(ct context.Context, appID uint64, versionNumber int32) ([]entity.Team, *errs.Error) {
	appVersionVisibleTeams, err := a.appVersionVisibleTeamDao.FindAppVersionVisibleTeamsByAppIDAndVersionNumber(ct, appID, versionNumber)
	if err != nil {
		return nil, err
	}

	teamIDs := collect.Map(appVersionVisibleTeams, func(appVersionVisibleTeam entity.AppVersionVisibleTeam, _ int) uint64 {
		return appVersionVisibleTeam.TeamID
	})
	teams, err := a.teamDao.FindTeamsByIDs(ct, teamIDs)
	if err != nil {
		return nil, err
	}

	return teams, nil
}

func (a App) CreateApp(ct context.Context, name string) (entity.App, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.App{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	genAppIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "appID"}
	genAppIDRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genAppIDReq)
	if rpcErr != nil {
		return entity.App{}, errs.FromGRPCErr(rpcErr)
	}

	genAppSecretReq := &proto.GenerateUniqueStringRequest{SequenceName: "apiSecret"}
	genAppSecretRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueString(ct, genAppSecretReq)
	if rpcErr != nil {
		return entity.App{}, errs.FromGRPCErr(rpcErr)
	}

	app := entity.App{
		ID:                genAppIDRes.UniqueNumber,
		Name:              name,
		Description:       "",
		APISecret:         genAppSecretRes.UniqueString,
		InstallationCount: 0,
		CreatorUserID:     userID,
		CreatedAt:         time.Now().UTC(),
	}
	err := a.appDao.CreateApp(ct, app)
	if err != nil {
		return entity.App{}, err
	}

	if feature.EnableAuthorization {
		err = a.authorizer.registerResource(ct, authorization.AppResourceType, app.ID)
		if err != nil {
			return entity.App{}, err
		}

		// When create a new app,
		// 1) Resource: register app resource
		// 2) UserGroup: create AppAdmin and AppMember userGroups
		// 3) UserGroupMember:
		// 		add app creator to AppAdmin group
		// 		add app creator to AppMember group
		// 4) Permissions:
		//		assign AppAdmin permissions to `AppAdmin` group
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

		_, err = a.authorizer.createUserGroupAndAssignPermissions(ct,
			userID,
			appAdminUserGroupName,
			&appAdminDescription,
			appAdminOperations,
		)
		if err != nil {
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
			return entity.App{}, err
		}
	}

	return app, nil
}

func (a App) UpdateApp(ct context.Context, appID uint64, input UpdateAppInput) (entity.App, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.App{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewUpdateAppQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.App{}, err
		}

		if !hasPermission {
			return entity.App{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("authorization query: %v", query))
		}
	}

	app, err := a.appDao.FindAppByID(ct, appID)
	if err != nil {
		return entity.App{}, err
	}

	if input.Name != nil {
		app.Name = *input.Name
	}
	if input.Description != nil {
		app.Description = *input.Description
	}
	if input.ActiveVersionNumber != nil {
		if app.ActiveVersionNumber != nil {
			if *app.ActiveVersionNumber > *input.ActiveVersionNumber {
				return entity.App{}, errs.NewError(errs.InvalidOperation, fmt.Sprintf(
					"Cannot rollback app version: appID=%v, prevAppVesion=%v newAppVersion=%v", appID, *app.ActiveVersionNumber, *input.ActiveVersionNumber))
			}

			if *app.ActiveVersionNumber < *input.ActiveVersionNumber {
				appVersion, internalErr := a.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, *input.ActiveVersionNumber)
				if internalErr != nil {
					return entity.App{}, internalErr
				}

				if !appVersion.IsPublic {
					return entity.App{}, errs.NewError(errs.InvalidOperation, fmt.Sprintf(
						"Cannot activate a non-public app version: appID=%v, appVersion=%v", appID, *input.ActiveVersionNumber))
				}

				// roll forward app installation automatically
				a.rollForwardAppInstallations(ct, appID, *input.ActiveVersionNumber)
			}
		}

		app.ActiveVersionNumber = input.ActiveVersionNumber
	}

	now := time.Now().UTC()
	app.UpdatedAt = &now
	err = a.appDao.UpdateApp(ct, app)
	if err != nil {
		return entity.App{}, err
	}

	return app, nil
}

func (a App) RefreshAppSecret(ct context.Context, appID uint64) (entity.App, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.App{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewRefreshAppSecretQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.App{}, err
		}

		if !hasPermission {
			return entity.App{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("authorization query: %v", query))
		}
	}

	app, err := a.appDao.FindAppByID(ct, appID)
	if err != nil {
		return entity.App{}, err
	}

	genAppSecretReq := &proto.GenerateUniqueStringRequest{SequenceName: "apiSecret"}
	genAppSecretRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueString(ct, genAppSecretReq)
	if rpcErr != nil {
		return entity.App{}, errs.FromGRPCErr(rpcErr)
	}

	app.APISecret = genAppSecretRes.UniqueString
	err = a.appDao.UpdateApp(ct, app)
	if err != nil {
		return entity.App{}, err
	}

	return app, nil
}

func (a App) DeleteApp(ct context.Context, appID uint64) (entity.App, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.App{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewDeleteAppQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.App{}, err
		}

		if !hasPermission {
			return entity.App{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("authorization query: %v", query))
		}
	}

	app, err := a.appDao.FindAppByID(ct, appID)
	if err != nil {
		return entity.App{}, err
	}

	err = a.appDao.DeleteApp(ct, appID)
	if err != nil {
		return entity.App{}, err
	}

	// TODO(yuhang): delete registered app resource and groups

	return app, nil
}

func (a App) CreateAppVersion(ct context.Context, appID uint64) (entity.AppVersion, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.AppVersion{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if feature.EnableAuthorization {
		query := authorization.NewCreateAppVersionQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			return entity.AppVersion{}, errs.NewError(
			    errs.PermissionDenied, 
			    fmt.Sprintf("permission denied: authorization query=%v", query))
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
			return entity.AppVersion{}, err
		}
	}

	av.VersionNumber = maxVersion + 1
	err = a.appVersionDao.CreateAppVersion(ct, av)
	if err != nil {
		return entity.AppVersion{}, err
	}

	return av, nil
}

func (a App) UpdateAppVersion(ct context.Context, appID uint64, versionNumber int32, input UpdateAppVersionInput) (entity.AppVersion, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.AppVersion{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewUpdateAppVersionQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			return entity.AppVersion{}, errs.NewError(
			    errs.PermissionDenied, 
			    fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	av, internalErr := a.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
	if internalErr != nil {
		return entity.AppVersion{}, internalErr
	}

	app, internalErr := a.appDao.FindAppByID(ct, appID)
	if internalErr != nil {
		return entity.AppVersion{}, internalErr
	}
	if app.ActiveVersionNumber != nil && *app.ActiveVersionNumber == versionNumber && !input.IsPublic {
		return entity.AppVersion{}, errs.NewError(errs.InvalidOperation, "cannot mark an activated version as non-public")
	}

	av.HasUIExtension = input.HasUIExtension
	av.IsPublic = input.IsPublic
	av.IconURL = input.IconURL
	av.Changes = input.Changes
	av.UIExtensionEntrypointPath = input.UIExtensionEntryPointPath
	now := time.Now().UTC()
	av.UpdateAt = &now
	internalErr = a.appVersionDao.UpdateAppVersion(ct, av)
	if internalErr != nil {
		return entity.AppVersion{}, internalErr
	}

	return av, nil
}

func (a App) DeleteAppVersion(ct context.Context, appID uint64, versionNumber int32) (entity.AppVersion, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.AppVersion{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewDeleteAppVersionQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			return entity.AppVersion{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	av, err := a.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
	if err != nil {
		return entity.AppVersion{}, err
	}

	app, err := a.appDao.FindAppByID(ct, appID)
	if err != nil {
		return entity.AppVersion{}, err
	}

	if app.ActiveVersionNumber != nil && *app.ActiveVersionNumber == versionNumber {
		return entity.AppVersion{}, errs.NewError(errs.InvalidOperation, fmt.Sprintf(
			"Cannot delete active version: appID=%v", appID))
	}

	// TODO(yuhang): below operations should be atomic by wrapping in one txn. We'll implement that after adding our
	// own transaction library.
	err = a.appVersionDao.DeleteAppVersion(ct, appID, versionNumber)
	if err != nil {
		return entity.AppVersion{}, err
	}

	err = a.appTeamInstallationDao.DeleteAppTeamInstallationsByAppIDAndVersionNumber(ct, appID, versionNumber)
	if err != nil {
		return entity.AppVersion{}, err
	}

	err = a.appVersionVisibleTeamDao.DeleteAppVersionVisibleTeamsByAppIDAndVersionNumber(ct, appID, versionNumber)
	if err != nil {
		return entity.AppVersion{}, err
	}

	return av, nil
}

func (a App) CreateAppVersionVisibleTeam(ct context.Context, appID uint64, versionNumber int32, teamID uint64) (entity.AppVersion, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.AppVersion{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewCreateAppVersionVisibleTeamQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			return entity.AppVersion{}, errs.NewError(
			    errs.PermissionDenied, 
			    fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	av := entity.AppVersionVisibleTeam{
		AppID:         appID,
		VersionNumber: versionNumber,
		TeamID:        teamID,
	}
	err := a.appVersionVisibleTeamDao.CreateAppVersionVisibleTeam(ct, av)
	if err != nil {
		return entity.AppVersion{}, err
	}

	appVersion, err := a.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
	if err != nil {
		return entity.AppVersion{}, err
	}

	return appVersion, nil
}

func (a App) DeleteAppVersionVisibleTeam(ct context.Context, appID uint64, versionNumber int32, teamID uint64) (entity.AppVersion, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.AppVersion{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewDeleteAppVersionVisibleTeamQuery(userID, appID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			return entity.AppVersion{}, errs.NewError(
			    errs.PermissionDenied, 
			    fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	av, err := a.appVersionVisibleTeamDao.FindAppVersionVisibleTeam(ct, appID, versionNumber, teamID)
	if err != nil {
		return entity.AppVersion{}, err
	}

	err = a.appVersionVisibleTeamDao.DeleteAppVersionVisibleTeam(ct, appID, versionNumber, teamID)
	if err != nil {
		return entity.AppVersion{}, err
	}

	// if team has installed the version, delete installation as well
	appTeamInstallation, err := a.appTeamInstallationDao.FindAppTeamInstallationByAppIDAndTeamID(ct, appID, teamID)
	if err != nil {
		if err.Code != errs.NotFound {
			return entity.AppVersion{}, err
		}
	} else {
		if appTeamInstallation.EnabledVersionNumber == versionNumber {
			err = a.appTeamInstallationDao.DeleteAppTeamInstallation(ct, appID, teamID)
			if err != nil {
				return entity.AppVersion{}, err
			}
		}
	}

	appVersion, err := a.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, av.AppID, av.VersionNumber)
	if err != nil {
		return entity.AppVersion{}, err
	}

	return appVersion, nil
}

func (a App) CreateAppInstallation(ct context.Context, teamID uint64, appID uint64, versionNumber int32) (entity.AppTeamInstallation, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.AppTeamInstallation{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if feature.EnableAuthorization {
		query := authorization.NewCreateAppTeamInstallationQuery(userID, teamID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.AppTeamInstallation{}, err
		}

		if !hasPermission {
			return entity.AppTeamInstallation{}, errs.NewError(
			     errs.PermissionDenied, 
			     fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	app, err := a.appDao.FindAppByID(ct, appID)
	if err != nil {
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
		return entity.AppTeamInstallation{}, err
	}

	app.InstallationCount = app.InstallationCount + 1
	err = a.appDao.UpdateApp(ct, app)
	if err != nil {
		return entity.AppTeamInstallation{}, err
	}

	return ai, nil
}

func (a App) UpdateAppInstallation(ct context.Context, appID uint64, teamID uint64, input UpdateAppTeamInstallationInput) (entity.AppTeamInstallation, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.AppTeamInstallation{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewUpdateAppTeamInstallationQuery(userID, teamID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.AppTeamInstallation{}, err
		}

		if !hasPermission {
			return entity.AppTeamInstallation{}, errs.NewError(
			    errs.PermissionDenied, 
			    fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	ai, err := a.appTeamInstallationDao.FindAppTeamInstallationByAppIDAndTeamID(ct, appID, teamID)
	if err != nil {
		return entity.AppTeamInstallation{}, err
	}

	ai.EnabledVersionNumber = input.EnabledVersionNumber
	err = a.appTeamInstallationDao.UpdateAppTeamInstallation(ct, ai)
	if err != nil {
		return entity.AppTeamInstallation{}, err
	}

	return ai, nil
}

func (a App) DeleteAppInstallation(ct context.Context, appID uint64, teamID uint64) (entity.AppTeamInstallation, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.AppTeamInstallation{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewDeleteAppTeamInstallationQuery(userID, teamID)
		hasPermission, err := a.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.AppTeamInstallation{}, err
		}

		if !hasPermission {
			return entity.AppTeamInstallation{}, errs.NewError(
			    errs.PermissionDenied, 
			    fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	ai, err := a.appTeamInstallationDao.FindAppTeamInstallationByAppIDAndTeamID(ct, appID, teamID)
	if err != nil {
		return entity.AppTeamInstallation{}, err
	}

	err = a.appTeamInstallationDao.DeleteAppTeamInstallation(ct, appID, teamID)
	if err != nil {
		return entity.AppTeamInstallation{}, err
	}

	return ai, nil
}

// rollForwardAppInstallations moves all app installations to a newly enabled version
func (a App) rollForwardAppInstallations(ct context.Context, appID uint64, activeVersionNumber int32) *errs.Error {
	appInstallations, err := a.appTeamInstallationDao.FindAppTeamInstallationsByAppID(ct, appID)
	if err != nil {
		return err
	}

	for _, appInstallation := range appInstallations {
		if appInstallation.EnabledVersionNumber < activeVersionNumber {
			appInstallation.EnabledVersionNumber = activeVersionNumber
			err = a.appTeamInstallationDao.UpdateAppTeamInstallation(ct, appInstallation)
			if err != nil {
				a.logger.ErrorWithContext(ct, err)
			}
		}
	}

	return nil
}

func (a App) isAppVisibleToTeam(ct context.Context, app entity.App, teamID uint64) (bool, *errs.Error) {
	appVersions, err := a.appVersionDao.FindAppVersionsByAppID(ct, app.ID)
	if err != nil {
		return false, err
	}

	appVersions = collect.Filter(appVersions, func(appVersion entity.AppVersion) bool {
		var visible bool
		visible, err = a.isAppVersionVisibleToTeam(ct, app, appVersion, teamID)
		return visible
	})

	if err != nil {
		return false, err
	}

	if len(appVersions) > 0 {
		return true, nil
	}

	return false, nil
}

func (a App) isAppVersionVisibleToTeam(ct context.Context, app entity.App, appVersion entity.AppVersion, teamID uint64) (bool, *errs.Error) {
	if app.ActiveVersionNumber != nil && appVersion.VersionNumber < *app.ActiveVersionNumber {
		// if active version has been set, we should filter all old versions
		return false, nil
	}

	if !appVersion.IsPublic {
		// if app version not public, we need to check if team is in visible list
		_, err := a.appVersionVisibleTeamDao.FindAppVersionVisibleTeam(ct, app.ID, appVersion.VersionNumber, teamID)
		if err != nil {
			if err.Code != errs.NotFound {
				return false, err
			}

			return false, nil
		}

		return true, nil
	}

	return true, nil
}

func NewApp(
	logger telemetry.Logger,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	authorizer Authorizer,
	appDao dao.App,
	appVersionDao dao.AppVersion,
	appTeamInstallationDao dao.AppTeamInstallation,
	appVersionVisibleTeamDao dao.AppVersionVisibleTeam,
	teamDao dao.Team,
) App {
	return App{
		logger,
		cloudClientRegistry,
		authorizer,
		appDao,
		appVersionDao,
		appTeamInstallationDao,
		appVersionVisibleTeamDao,
		teamDao,
	}
}
