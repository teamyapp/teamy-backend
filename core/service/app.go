package service

import (
	"context"
	"fmt"
	"time"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/client"
	cloudAuthorization "github.com/teamyapp/cloud/libs/authorization"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type App struct {
	logger                     telemetry.Logger
	cloudClientRegistry        *client.Registry
	authorizer                 client.Authorizer
	featureToggles             feature.Toggles
	transactionFactory         transaction.Factory
	stateSyncer                *realtime.StateSyncer
	appDaoV2                   daov2.App
	appVersionDaoV2            daov2.AppVersion
	appTeamInstallationDaoV2   daov2.AppTeamInstallation
	appVersionVisibleTeamDaoV2 daov2.AppVersionVisibleTeam
	teamDaoV2                  daov2.Team
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
	return a.appDaoV2.FindAppByID(ct, appID)
}

func (a App) FindApps(ct context.Context, filter *AppFilter) ([]entity.App, *errs.Error) {
	var apps []entity.App
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(true, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		if filter != nil {
			if filter.AppID != nil {
				app, err := a.appDaoV2.FindAppByIDWithTx(ct, tx, *filter.AppID)
				if err != nil {
					return err
				}

				apps = append(apps, app)
			} else {
				var err *errs.Error
				apps, err = a.appDaoV2.FindAllAppsWithTx(ct, tx)
				if err != nil {
					return err
				}
			}

			if filter.TeamID != nil {
				// find all apps that are visible to the team
				var filtered []entity.App
				for _, app := range apps {
					visible, err := a.isAppVisibleToTeam(ct, tx, app, *filter.TeamID)
					if err != nil {
						return err
					}

					if visible {
						filtered = append(filtered, app)
					}
				}

				apps = filtered
			}
		} else {
			var err *errs.Error
			apps, err = a.appDaoV2.FindAllAppsWithTx(ct, tx)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return apps, nil
}

func (a App) FindAppTeamInstallationsByAppID(ct context.Context, appID uint64) ([]entity.AppTeamInstallation, *errs.Error) {
	return a.appTeamInstallationDaoV2.FindAppTeamInstallationsByAppID(ct, appID)
}

func (a App) FindAppInstallationsByTeamID(ct context.Context, teamID uint64) ([]entity.AppTeamInstallation, *errs.Error) {
	return a.appTeamInstallationDaoV2.FindAppTeamInstallationsByTeamID(ct, teamID)
}

func (a App) FindAppVersionByAppID(ct context.Context, appID uint64) ([]entity.AppVersion, *errs.Error) {
	return a.appVersionDaoV2.FindAppVersionsByAppID(ct, appID)
}

func (a App) FindAppVersionByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int32) (entity.AppVersion, *errs.Error) {
	return a.appVersionDaoV2.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
}

func (a App) FindAppVersionVisibleTeams(ct context.Context, appID uint64, versionNumber int32) ([]entity.Team, *errs.Error) {
	var teams []entity.Team
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(true, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		appVersionVisibleTeams, err :=
			a.appVersionVisibleTeamDaoV2.FindAppVersionVisibleTeamsByAppIDAndVersionNumberWithTx(ct, tx, appID, versionNumber)
		if err != nil {
			return err
		}

		teamIDs := collect.Map(appVersionVisibleTeams, func(appVersionVisibleTeam entity.AppVersionVisibleTeam, _ int) uint64 {
			return appVersionVisibleTeam.TeamID
		})
		teams, err = a.teamDaoV2.FindTeamsByIDsWithTx(ct, tx, teamIDs)
		if err != nil {
			return err
		}

		return nil
	})

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
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		return a.appDaoV2.CreateApp(ct, tx, app)
	})

	if err != nil {
		return entity.App{}, err
	}

	if a.featureToggles.EnableAuthorization {
		err = a.authorizer.RegisterResource(ct, authorization.AppResourceType, app.ID)
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
		appAdminOperations := make([]cloudAuthorization.ResourceOperation, 0)
		for _, appAdminResourceTypeOperation := range authorization.AppAdminResourceTypeOperations {
			appAdminOperations = append(appAdminOperations, cloudAuthorization.ResourceOperation{
				ResourceType: appAdminResourceTypeOperation.ResourceType,
				Operation:    appAdminResourceTypeOperation.Operation,
				ResourceID:   app.ID,
			})
		}

		_, err = a.authorizer.CreateUserGroupAndAssignPermissions(ct,
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
		appMemberOperations := make([]cloudAuthorization.ResourceOperation, 0)
		for _, appMemberResourceTypeOperation := range authorization.AppMemberResourceTypeOperations {
			appMemberOperations = append(appMemberOperations, cloudAuthorization.ResourceOperation{
				ResourceType: appMemberResourceTypeOperation.ResourceType,
				Operation:    appMemberResourceTypeOperation.Operation,
				ResourceID:   app.ID,
			})
		}

		_, err = a.authorizer.CreateUserGroupAndAssignPermissions(ct,
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
	if a.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.App{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewUpdateInAppQuery(userID, appID)
		hasPermission, err := a.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.App{}, err
		}

		if !hasPermission {
			return entity.App{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("authorization query: %v", query))
		}
	}

	var app entity.App
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		app, internalErr = a.appDaoV2.FindAppByIDWithTx(ct, tx, appID)
		if internalErr != nil {
			return internalErr
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
					return errs.NewError(errs.InvalidOperation, fmt.Sprintf(
						"Cannot rollback app version: appID=%v, prevAppVesion=%v newAppVersion=%v", appID, *app.ActiveVersionNumber, *input.ActiveVersionNumber))
				}

				if *app.ActiveVersionNumber < *input.ActiveVersionNumber {
					var appVersion entity.AppVersion
					appVersion, internalErr = a.appVersionDaoV2.FindAppVersionByAppIDAndVersionNumberWithTx(ct, tx,
						appID, *input.ActiveVersionNumber)
					if internalErr != nil {
						return internalErr
					}

					if !appVersion.IsPublic {
						return errs.NewError(errs.InvalidOperation, fmt.Sprintf(
							"Cannot activate a non-public app version: appID=%v, appVersion=%v", appID, *input.ActiveVersionNumber))
					}

					// roll forward app installation automatically
					a.rollForwardAppInstallations(ct, tx, appID, *input.ActiveVersionNumber)
				}
			}

			app.ActiveVersionNumber = input.ActiveVersionNumber
		}

		now := time.Now().UTC()
		app.UpdatedAt = &now
		internalErr = a.appDaoV2.UpdateApp(ct, tx, app)
		if internalErr != nil {
			return internalErr
		}

		return nil
	})

	if err != nil {
		return entity.App{}, err
	}

	return app, nil
}

func (a App) RefreshAppSecret(ct context.Context, appID uint64) (entity.App, *errs.Error) {
	if a.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.App{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewRefreshAppSecretInAppQuery(userID, appID)
		hasPermission, err := a.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.App{}, err
		}

		if !hasPermission {
			return entity.App{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("authorization query: %v", query))
		}
	}

	genAppSecretReq := &proto.GenerateUniqueStringRequest{SequenceName: "apiSecret"}
	genAppSecretRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueString(ct, genAppSecretReq)
	if rpcErr != nil {
		return entity.App{}, errs.FromGRPCErr(rpcErr)
	}

	var app entity.App
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		app, internalErr = a.appDaoV2.FindAppByIDWithTx(ct, tx, appID)
		if internalErr != nil {
			return internalErr
		}

		app.APISecret = genAppSecretRes.UniqueString
		internalErr = a.appDaoV2.UpdateApp(ct, tx, app)
		if internalErr != nil {
			return internalErr
		}

		return nil
	})

	if err != nil {
		return entity.App{}, err
	}

	return app, nil
}

func (a App) DeleteApp(ct context.Context, appID uint64) (entity.App, *errs.Error) {
	if a.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.App{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewDeleteInAppQuery(userID, appID)
		hasPermission, err := a.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.App{}, err
		}

		if !hasPermission {
			return entity.App{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("authorization query: %v", query))
		}
	}

	var app entity.App
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		app, internalErr = a.appDaoV2.FindAppByIDWithTx(ct, tx, appID)
		if internalErr != nil {
			return internalErr
		}

		internalErr = a.appDaoV2.DeleteApp(ct, tx, appID)
		if internalErr != nil {
			return internalErr
		}

		return nil
	})

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

	if a.featureToggles.EnableAuthorization {
		query := authorization.NewCreateAppVersionInAppQuery(userID, appID)
		hasPermission, err := a.authorizer.HasPermission(ct, query)
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
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		maxVersion, err := a.appVersionDaoV2.FindMaxVersionNumberWithTx(ct, tx, appID)
		if err != nil {
			if err.Code == errs.NotFound {
				// no version exists, start from 0
				maxVersion = 0
			} else {
				return err
			}
		}

		av.VersionNumber = maxVersion + 1
		err = a.appVersionDaoV2.CreateAppVersion(ct, tx, av)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return entity.AppVersion{}, err
	}

	return av, nil
}

func (a App) UpdateAppVersion(ct context.Context, appID uint64, versionNumber int32, input UpdateAppVersionInput) (entity.AppVersion, *errs.Error) {
	if a.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.AppVersion{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewUpdateAppVersionInAppQuery(userID, appID)
		hasPermission, err := a.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			return entity.AppVersion{}, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var av entity.AppVersion
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		av, internalErr = a.appVersionDaoV2.FindAppVersionByAppIDAndVersionNumberWithTx(ct, tx, appID, versionNumber)
		if internalErr != nil {
			return internalErr
		}

		app, internalErr := a.appDaoV2.FindAppByIDWithTx(ct, tx, appID)
		if internalErr != nil {
			return internalErr
		}

		if app.ActiveVersionNumber != nil &&
			*app.ActiveVersionNumber == versionNumber &&
			!input.IsPublic {
			return errs.NewError(errs.InvalidOperation, "cannot mark an activated version as non-public")
		}

		av.HasUIExtension = input.HasUIExtension
		av.IsPublic = input.IsPublic
		av.IconURL = input.IconURL
		av.Changes = input.Changes
		av.UIExtensionEntrypointPath = input.UIExtensionEntryPointPath
		now := time.Now().UTC()
		av.UpdateAt = &now
		internalErr = a.appVersionDaoV2.UpdateAppVersion(ct, tx, av)
		if internalErr != nil {
			return internalErr
		}

		return nil
	})

	if err != nil {
		return entity.AppVersion{}, err
	}

	return av, nil
}

func (a App) DeleteAppVersion(ct context.Context, appID uint64, versionNumber int32) (entity.AppVersion, *errs.Error) {
	if a.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.AppVersion{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewDeleteAppVersionInAppQuery(userID, appID)
		hasPermission, err := a.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			return entity.AppVersion{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var av entity.AppVersion
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		av, internalErr = a.appVersionDaoV2.FindAppVersionByAppIDAndVersionNumberWithTx(ct, tx, appID, versionNumber)
		if internalErr != nil {
			return internalErr
		}

		app, err := a.appDaoV2.FindAppByIDWithTx(ct, tx, appID)
		if err != nil {
			return internalErr
		}

		if app.ActiveVersionNumber != nil && *app.ActiveVersionNumber == versionNumber {
			return errs.NewError(errs.InvalidOperation, fmt.Sprintf("Cannot delete active version: appID=%v", appID))
		}

		internalErr = a.appVersionDaoV2.DeleteAppVersion(ct, tx, appID, versionNumber)
		if internalErr != nil {
			return internalErr
		}

		internalErr = a.appTeamInstallationDaoV2.DeleteAppTeamInstallationsByAppIDAndVersionNumber(ct, tx, appID,
			versionNumber)
		if internalErr != nil {
			return internalErr
		}

		internalErr = a.appVersionVisibleTeamDaoV2.DeleteAppVersionVisibleTeamsByAppIDAndVersionNumber(ct, tx, appID,
			versionNumber)
		if internalErr != nil {
			return internalErr
		}

		return nil
	})

	if err != nil {
		return entity.AppVersion{}, err
	}

	return av, nil
}

func (a App) CreateAppVersionVisibleTeam(ct context.Context, appID uint64, versionNumber int32, teamID uint64) (entity.AppVersion, *errs.Error) {
	if a.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.AppVersion{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewCreateAppVersionVisibleTeamInAppQuery(userID, appID)
		hasPermission, err := a.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			return entity.AppVersion{}, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var appVersion entity.AppVersion
	av := entity.AppVersionVisibleTeam{
		AppID:         appID,
		VersionNumber: versionNumber,
		TeamID:        teamID,
	}
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		err := a.appVersionVisibleTeamDaoV2.CreateAppVersionVisibleTeam(ct, tx, av)
		if err != nil {
			return err
		}

		appVersion, err = a.appVersionDaoV2.FindAppVersionByAppIDAndVersionNumberWithTx(ct, tx, appID, versionNumber)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return entity.AppVersion{}, err
	}

	return appVersion, nil
}

func (a App) DeleteAppVersionVisibleTeam(ct context.Context, appID uint64, versionNumber int32, teamID uint64) (entity.AppVersion, *errs.Error) {
	if a.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.AppVersion{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewDeleteAppVersionVisibleTeamInAppQuery(userID, appID)
		hasPermission, err := a.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.AppVersion{}, err
		}

		if !hasPermission {
			return entity.AppVersion{}, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var appVersion entity.AppVersion
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		av, err := a.appVersionVisibleTeamDaoV2.FindAppVersionVisibleTeamWithTx(ct, tx, appID, versionNumber, teamID)
		if err != nil {
			return err
		}

		err = a.appVersionVisibleTeamDaoV2.DeleteAppVersionVisibleTeam(ct, tx, appID, versionNumber, teamID)
		if err != nil {
			return err
		}

		// if team has installed the version, delete installation as well
		appTeamInstallation, err := a.appTeamInstallationDaoV2.FindAppTeamInstallationByAppIDAndTeamIDWithTx(ct, tx,
			appID, teamID)
		if err != nil {
			if err.Code != errs.NotFound {
				return err
			}
		} else {
			if appTeamInstallation.EnabledVersionNumber == versionNumber {
				err = a.appTeamInstallationDaoV2.DeleteAppTeamInstallation(ct, tx, appID, teamID)
				if err != nil {
					return err
				}
			}
		}

		appVersion, err = a.appVersionDaoV2.FindAppVersionByAppIDAndVersionNumberWithTx(ct, tx, av.AppID, av.VersionNumber)
		if err != nil {
			return err
		}

		return nil
	})

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

	if a.featureToggles.EnableAuthorization {
		query := authorization.NewUpdateAppInstallationInTeamQuery(userID, teamID)
		hasPermission, err := a.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.AppTeamInstallation{}, err
		}

		if !hasPermission {
			return entity.AppTeamInstallation{}, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	ai := entity.AppTeamInstallation{
		AppID:                appID,
		InstalledTeamID:      teamID,
		InstalledByUserID:    &userID,
		EnabledVersionNumber: versionNumber,
		InstalledAt:          time.Now().UTC(),
	}
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		app, err := a.appDaoV2.FindAppByIDWithTx(ct, tx, appID)
		if err != nil {
			return err
		}

		err = a.appTeamInstallationDaoV2.CreateAppTeamInstallation(ct, tx, ai)
		if err != nil {
			return err
		}

		app.InstallationCount = app.InstallationCount + 1
		err = a.appDaoV2.UpdateApp(ct, tx, app)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return entity.AppTeamInstallation{}, err
	}

	return ai, nil
}

func (a App) UpdateAppInstallation(ct context.Context, appID uint64, teamID uint64, input UpdateAppTeamInstallationInput) (entity.AppTeamInstallation, *errs.Error) {
	if a.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.AppTeamInstallation{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewUpdateAppInstallationInTeamQuery(userID, teamID)
		hasPermission, err := a.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.AppTeamInstallation{}, err
		}

		if !hasPermission {
			return entity.AppTeamInstallation{}, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var ai entity.AppTeamInstallation
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		ai, err = a.appTeamInstallationDaoV2.FindAppTeamInstallationByAppIDAndTeamIDWithTx(ct, tx, appID, teamID)
		if err != nil {
			return err
		}

		ai.EnabledVersionNumber = input.EnabledVersionNumber
		err = a.appTeamInstallationDaoV2.UpdateAppTeamInstallation(ct, tx, ai)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return entity.AppTeamInstallation{}, err
	}

	return ai, nil
}

func (a App) DeleteAppInstallation(ct context.Context, appID uint64, teamID uint64) (entity.AppTeamInstallation, *errs.Error) {
	if a.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.AppTeamInstallation{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewUpdateAppInstallationInTeamQuery(userID, teamID)
		hasPermission, err := a.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.AppTeamInstallation{}, err
		}

		if !hasPermission {
			return entity.AppTeamInstallation{}, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var ai entity.AppTeamInstallation
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		ai, err = a.appTeamInstallationDaoV2.FindAppTeamInstallationByAppIDAndTeamIDWithTx(ct, tx, appID, teamID)
		if err != nil {
			return err
		}

		err = a.appTeamInstallationDaoV2.DeleteAppTeamInstallation(ct, tx, appID, teamID)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return entity.AppTeamInstallation{}, err
	}

	return ai, nil
}

// rollForwardAppInstallations moves all app installations to a newly enabled version
func (a App) rollForwardAppInstallations(ct context.Context, tx *transaction.Transaction, appID uint64, activeVersionNumber int32) *errs.Error {
	appInstallations, err := a.appTeamInstallationDaoV2.FindAppTeamInstallationsByAppIDWithTx(ct, tx, appID)
	if err != nil {
		return err
	}

	for _, appInstallation := range appInstallations {
		if appInstallation.EnabledVersionNumber < activeVersionNumber {
			appInstallation.EnabledVersionNumber = activeVersionNumber
			err = a.appTeamInstallationDaoV2.UpdateAppTeamInstallation(ct, tx, appInstallation)
			if err != nil {
				a.logger.ErrorWithContext(ct, err)
			}
		}
	}

	return nil
}

func (a App) isAppVisibleToTeam(
	ct context.Context,
	tx *transaction.Transaction,
	app entity.App,
	teamID uint64,
) (bool, *errs.Error) {
	appVersions, err := a.appVersionDaoV2.FindAppVersionsByAppIDWithTx(ct, tx, app.ID)
	if err != nil {
		return false, err
	}

	var filtered []entity.AppVersion
	for _, appVersion := range appVersions {
		var visible bool
		visible, err = a.isAppVersionVisibleToTeam(ct, tx, app, appVersion, teamID)
		if err != nil {
			return false, err
		}

		if visible {
			filtered = append(filtered, appVersion)
		}
	}

	if len(filtered) > 0 {
		return true, nil
	}

	return false, nil
}

func (a App) isAppVersionVisibleToTeam(
	ct context.Context,
	tx *transaction.Transaction,
	app entity.App,
	appVersion entity.AppVersion,
	teamID uint64,
) (bool, *errs.Error) {
	if app.ActiveVersionNumber != nil && appVersion.VersionNumber < *app.ActiveVersionNumber {
		// if active version has been set, we should filter all old versions
		return false, nil
	}

	if !appVersion.IsPublic {
		// if app version not public, we need to check if team is in visible list
		_, err := a.appVersionVisibleTeamDaoV2.FindAppVersionVisibleTeamWithTx(ct, tx, app.ID, appVersion.VersionNumber,
			teamID)
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
	cloudClientRegistry *client.Registry,
	authorizer client.Authorizer,
	featureToggles feature.Toggles,
	transactionFactory transaction.Factory,
	stateSyncer *realtime.StateSyncer,
	appDaoV2 daov2.App,
	appVersionDaoV2 daov2.AppVersion,
	appTeamInstallationDaoV2 daov2.AppTeamInstallation,
	appVersionVisibleTeamDaoV2 daov2.AppVersionVisibleTeam,
	teamDaoV2 daov2.Team,
) App {
	return App{
		logger,
		cloudClientRegistry,
		authorizer,
		featureToggles,
		transactionFactory,
		stateSyncer,
		appDaoV2,
		appVersionDaoV2,
		appTeamInstallationDaoV2,
		appVersionVisibleTeamDaoV2,
		teamDaoV2,
	}
}
