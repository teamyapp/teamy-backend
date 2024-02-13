package service

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/client"
	cloudAuthorization "github.com/teamyapp/cloud/libs/authorization"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	tmio "github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/cloud/libs/randgen"
	"github.com/teamyapp/cloud/libs/security"
	"github.com/teamyapp/cloud/libs/storage"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudTransaction "github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/repository"
	"github.com/teamyapp/teamy-backend/core/transaction"
	"google.golang.org/protobuf/types/known/emptypb"
	"gopkg.in/yaml.v3"
)

var appPackageRoot = path.Join("app", "packages")
var secretAlphabet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!?@#_-")
var secretLength = 32
var defaultAppGroupName = "default app group"
var defaultRolloutName = "default app rollout"
var defaultAppVersionNumber = 1

type CreateAppInput struct {
	Name string
}
type UploadFunc func(io.Reader) *errs.Error

const (
	readAheadBytes = 28
	maxReadBytes   = 4096
	bufferSize     = maxReadBytes + readAheadBytes
)

type App struct {
	logger                     telemetry.Logger
	storageMapClient           storage.MapClient
	cloudClientRegistry        *client.Registry
	authorizer                 client.Authorizer
	featureToggles             feature.Toggles
	transactionFactory         cloudTransaction.Factory
	stateSyncer                *realtime.StateSyncer
	appDao                     dao.App
	appVersionDao              dao.AppVersion
	appVersionPriceDao         dao.AppVersionPrice
	appVersionChangeDao        dao.AppVersionChange
	appSecretDao               dao.AppSecret
	appPackageUploadSessionDao dao.AppPackageUploadSession
	teamAppInstallationDao     dao.TeamAppInstallation
	teamDao                    dao.Team
	tagDao                     dao.Tag
	appTagRelationDao          dao.AppTagRelation
	appGroupRelationDao        dao.AppGroupRelation
	groupRolloutRelationDao    dao.GroupRolloutRelation
	groupMemberRelationDao     dao.GroupMemberRelation
	appRolloutRelation         dao.AppRolloutRelation
	rolloutDao                 dao.Rollout
	groupRepository            *repository.Group
	activatorRepository        *repository.Activator
	versionSelectorRepository  *repository.VersionSelector
	jwtAuthority               security.JWTAuthority
}

type AppFilter struct {
	AppID  *uint64
	TeamID *uint64
}

type UpdateAppTeamInstallationInput struct {
	EnabledVersionNumber int32
}

type CreateAppVersionInput struct {
	AppName     string
	Description string
}

type CreateAppSecretInput struct {
	Name string
}

type UpdateAppSecretInput struct {
	Name string
}

type UpdateAppInput struct {
	Tags []string
}

type GenerateTokenInput struct {
	SecretID uint64
	Secret   string
}

func (a App) FindAppByID(ct context.Context, appID uint64) (entity.App, *errs.Error) {
	return a.appDao.FindAppByID(ct, appID)
}

func (a App) FindAppsByGroupID(ct context.Context, groupID uint64) ([]entity.App, *errs.Error) {
	txCtx := transaction.NewTransactionsContext(
		a.logger,
		a.transactionFactory,
		a.stateSyncer,
		ct,
	)

	var apps []entity.App
	err := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		appIDs, err := a.appGroupRelationDao.FindAppIDsByGroupIDWithTx(ct, tx, groupID)
		if err != nil {
			return err
		}

		if len(appIDs) == 0 {
			return nil
		}

		apps, err = a.appDao.FindAppsByAppIDsWithTx(ct, tx, appIDs)
		return err
	})

	return apps, err
}

func (a App) FindSecretsByAppID(ct context.Context, appID uint64) ([]entity.AppSecret, *errs.Error) {
	return a.appSecretDao.FindSecretsByAppID(ct, appID)
}

func (a App) CreateAppSecret(ct context.Context, appID uint64, input CreateAppSecretInput) (entity.AppSecret, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.AppSecret{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	genAppSecretIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "appSecretID"}
	genAppSecretIDRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genAppSecretIDReq)
	if rpcErr != nil {
		return entity.AppSecret{}, errs.FromGRPCErr(rpcErr)
	}

	secretID := genAppSecretIDRes.UniqueNumber
	secret := randgen.String(secretAlphabet, secretLength)
	generateTokenInput := GenerateTokenInput{
		SecretID: secretID,
		Secret:   secret,
	}
	token, err := a.GetAppSecretToken(ct, generateTokenInput)
	if err != nil {
		return entity.AppSecret{}, err
	}

	appSecret := entity.AppSecret{
		ID:            secretID,
		Token:         token,
		Name:          input.Name,
		AppID:         appID,
		AddedAt:       time.Now().UTC(),
		AddedByUserID: userID,
	}
	txCtx := transaction.NewTransactionsContext(
		a.logger,
		a.transactionFactory,
		a.stateSyncer,
		ct,
	)
	err = txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		return a.appSecretDao.CreateAppSecret(ct, tx, appSecret)
	})
	return appSecret, err
}

func (a App) UpdateAppSecret(ct context.Context, appSecretID uint64, input UpdateAppSecretInput) (entity.AppSecret, *errs.Error) {
	var appSecret entity.AppSecret
	txCtx := transaction.NewTransactionsContext(
		a.logger,
		a.transactionFactory,
		a.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		appSecret, internalErr = a.appSecretDao.FindAppSecretByIDWithTx(ct, tx, appSecretID)
		if internalErr != nil {
			return internalErr
		}

		appSecret.Name = input.Name
		return a.appSecretDao.UpdateAppSecret(ct, tx, appSecretID, appSecret)
	})
	return appSecret, err
}

func (a App) DeleteAppSecret(ct context.Context, appSecretID uint64) (entity.AppSecret, *errs.Error) {
	var appSecret entity.AppSecret
	txCtx := transaction.NewTransactionsContext(
		a.logger,
		a.transactionFactory,
		a.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		appSecret, internalErr = a.appSecretDao.FindAppSecretByIDWithTx(ct, tx, appSecretID)
		if internalErr != nil {
			return internalErr
		}

		return a.appSecretDao.DeleteAppSecret(ct, tx, appSecretID)
	})
	return appSecret, err
}

func (a App) InstallAppToTeam(ct context.Context, appID uint64, teamID uint64) (entity.TeamAppInstallation, *errs.Error) {
	genTeamAppInstallationIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "teamAppInstallationID"}
	genTeamAppInstallationIDRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genTeamAppInstallationIDReq)
	if rpcErr != nil {
		return entity.TeamAppInstallation{}, errs.FromGRPCErr(rpcErr)
	}

	teamAppInstallation := entity.TeamAppInstallation{
		ID:              genTeamAppInstallationIDRes.UniqueNumber,
		AppID:           appID,
		InstalledTeamID: teamID,
	}

	txCtx := transaction.NewTransactionsContext(
		a.logger,
		a.transactionFactory,
		a.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		return a.teamAppInstallationDao.CreateTeamAppInstallation(ct, tx, teamAppInstallation)
	})
	return teamAppInstallation, err
}

func (a App) UninstallAppFromTeam(ct context.Context, appInstallationID uint64) (entity.TeamAppInstallation, *errs.Error) {
	var teamAppInstallation entity.TeamAppInstallation
	txCtx := transaction.NewTransactionsContext(
		a.logger,
		a.transactionFactory,
		a.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		teamAppInstallation, internalErr = a.teamAppInstallationDao.FindTeamAppInstallationByIDWithTx(ct, tx, appInstallationID)
		if internalErr != nil {
			return internalErr
		}

		return a.teamAppInstallationDao.DeleteTeamAppInstallationByID(ct, tx, appInstallationID)
	})
	return teamAppInstallation, err
}

func (a App) FindTeamAppInstallationsByAppID(ct context.Context, appID uint64) ([]entity.TeamAppInstallation, *errs.Error) {
	return a.teamAppInstallationDao.FindTeamAppInstallationsByAppID(ct, appID)
}

func (a App) FindTeamAppInstallationsByTeamID(ct context.Context, teamID uint64) ([]entity.TeamAppInstallation, *errs.Error) {
	return a.teamAppInstallationDao.FindTeamAppInstallationsByTeamID(ct, teamID)
}

func (a App) FindAppsByManagedByTeamID(ct context.Context, teamID uint64) ([]entity.App, *errs.Error) {
	return a.appDao.FindAppsByManagedByTeamID(ct, teamID)
}

func (a App) FindAppVersionByAppIDAndNumber(ct context.Context, appID uint64, versionNumber int) (entity.AppVersion, *errs.Error) {
	return a.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
}

func (a App) CreateApp(ct context.Context, teamID uint64, createAppInput CreateAppInput) (entity.App, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.App{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	genAppIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "appID"}
	genAppIDRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genAppIDReq)
	if rpcErr != nil {
		return entity.App{}, errs.FromGRPCErr(rpcErr)
	}

	now := time.Now().UTC()
	app := entity.App{
		ID:                 genAppIDRes.UniqueNumber,
		TotalInstallations: 0,
		ManagedByTeamID:    teamID,
		CreatedAt:          now,
	}

	appVersion := entity.AppVersion{
		AppID:           app.ID,
		Number:          defaultAppVersionNumber,
		AppName:         createAppInput.Name,
		Description:     "",
		CreatedByUserID: userID,
		IsReady:         true,
		Locked:          true,
		CreatedAt:       now,
	}

	genGroupIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "groupID"}
	genGroupIDRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genGroupIDReq)
	if rpcErr != nil {
		return entity.App{}, errs.FromGRPCErr(rpcErr)
	}

	staticTeamGroup := entity.StaticGroup{
		Group: entity.Group{
			ID:              genGroupIDRes.UniqueNumber,
			Type:            entity.GroupTypeStatic,
			MemberType:      entity.GroupMemberTypeTeam,
			Name:            defaultAppGroupName,
			MaxRolloutIndex: 0,
			CreatedAt:       now,
			Locked:          true,
		},
		MemberIDs: []uint64{teamID},
	}

	genActivatorIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "activatorID"}
	genActivatorIDRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genActivatorIDReq)
	if rpcErr != nil {
		return entity.App{}, errs.FromGRPCErr(rpcErr)
	}

	staticActivator := entity.StaticActivator{
		Activator: entity.Activator{
			ID:        genActivatorIDRes.UniqueNumber,
			Type:      entity.ActivatorTypeStatic,
			CreatedAt: now,
			Locked:    true,
		},
	}

	genVersionSelectorIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "versionSelectorID"}
	genVersionSelectorIDRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genVersionSelectorIDReq)
	if rpcErr != nil {
		return entity.App{}, errs.FromGRPCErr(rpcErr)
	}

	staticVersionSelector := entity.StaticVersionSelector{
		VersionSelector: entity.VersionSelector{
			ID:        genVersionSelectorIDRes.UniqueNumber,
			Type:      entity.VersionSelectorTypeStatic,
			CreatedAt: now,
			Locked:    true,
		},
		VersionNumber: defaultAppVersionNumber,
	}

	genRolloutIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "rolloutID"}
	genRolloutIDRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genRolloutIDReq)
	if rpcErr != nil {
		return entity.App{}, errs.FromGRPCErr(rpcErr)
	}

	rollout := entity.Rollout{
		ID:          genRolloutIDRes.UniqueNumber,
		Name:        defaultRolloutName,
		ActivatorID: staticActivator.ID,
		SelectorID:  staticVersionSelector.ID,
		Viewers:     1,
		IsEnabled:   true,
		Locked:      true,
		CreatedAt:   now,
	}

	txCtx := transaction.NewTransactionsContext(
		a.logger,
		a.transactionFactory,
		a.stateSyncer,
		ct,
	)
	internalErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		_, err := a.teamDao.FindTeamByIDWithTx(ct, tx, teamID)
		if err != nil {
			return err
		}

		err = a.appDao.CreateApp(ct, tx, app)
		if err != nil {
			return err
		}

		err = a.appVersionDao.CreateAppVersion(ct, tx, appVersion)
		if err != nil {
			return err
		}

		err = a.groupRepository.CreateStaticGroup(ct, tx, staticTeamGroup)
		if err != nil {
			return err
		}

		err = a.activatorRepository.CreateStaticActivator(ct, tx, staticActivator)
		if err != nil {
			return err
		}

		err = a.versionSelectorRepository.CreateStaticVersionSelector(ct, tx, staticVersionSelector)
		if err != nil {
			return err
		}

		err = a.rolloutDao.CreateRollout(ct, tx, rollout)
		if err != nil {
			return err
		}

		err = a.groupRolloutRelationDao.CreateGroupRolloutRelation(ct, tx, entity.GroupRolloutRelation{
			GroupID:    staticTeamGroup.ID,
			RolloutID:  rollout.ID,
			OrderIndex: 0,
		})
		if err != nil {
			return err
		}

		err = a.appGroupRelationDao.CreateAppGroupRelation(ct, tx, entity.AppGroupRelation{
			AppID:   app.ID,
			GroupID: staticTeamGroup.ID,
		})
		if err != nil {
			return err
		}

		return a.appRolloutRelation.CreateAppRolloutRelation(ct, tx, entity.AppRolloutRelation{
			AppID:     app.ID,
			RolloutID: rollout.ID,
			Type:      entity.AppRolloutRelationTypeTeam,
		})
	})

	if internalErr != nil {
		return entity.App{}, internalErr
	}

	if a.featureToggles.EnableAuthorization {
		internalErr = a.authorizer.RegisterResource(ct, authorization.AppResourceType, app.ID)
		if internalErr != nil {
			return entity.App{}, internalErr
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

		_, internalErr = a.authorizer.CreateUserGroupAndAssignPermissions(ct,
			userID,
			appAdminUserGroupName,
			&appAdminDescription,
			appAdminOperations,
		)
		if internalErr != nil {
			return entity.App{}, internalErr
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

		_, internalErr = a.authorizer.CreateUserGroupAndAssignPermissions(ct,
			userID,
			appMemberUserGroupName,
			&appMemberDescription,
			appMemberOperations,
		)
		if internalErr != nil {
			return entity.App{}, internalErr
		}
	}

	return app, nil
}

func (a App) FindTagsByAppID(ct context.Context, appID uint64) ([]entity.Tag, *errs.Error) {
	txCtx := transaction.NewTransactionsContext(
		a.logger,
		a.transactionFactory,
		a.stateSyncer,
		ct,
	)

	var tags []entity.Tag
	err := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		tagIDs, err := a.appTagRelationDao.FindTagIDsByAppIDWithTx(ct, tx, appID)
		if err != nil {
			return err
		}

		if len(tagIDs) == 0 {
			return nil
		}

		tags, err = a.tagDao.FindTagsByTagIDsWithTx(ct, tx, tagIDs)
		if err != nil {
			return err
		}

		return nil
	})

	return tags, err
}

func (a App) AddTagToApp(ct context.Context, appID uint64, value string) (entity.Tag, *errs.Error) {
	var tag entity.Tag
	txCtx := transaction.NewTransactionsContext(
		a.logger,
		a.transactionFactory,
		a.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		tag, internalErr = a.tagDao.FindTagByValueWithTx(ct, tx, value)
		if internalErr != nil {
			if internalErr.Code != errs.NotFound {
				return internalErr
			}

			genTagIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "tagID"}
			genTagIDRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genTagIDReq)
			if rpcErr != nil {
				return errs.FromGRPCErr(rpcErr)
			}

			tag = entity.Tag{
				ID:    genTagIDRes.UniqueNumber,
				Value: value,
			}

			internalErr = a.tagDao.CreateTag(ct, tx, tag)
			if internalErr != nil {
				return internalErr
			}

			appTagRelation := entity.AppTagRelation{
				AppID: appID,
				TagID: tag.ID,
			}

			return a.appTagRelationDao.CreateAppTagRelation(ct, tx, appTagRelation)
		}

		appTagRelation, internalErr := a.appTagRelationDao.FindAppTagByAppIDAndTagIDRelationWithTx(ct, tx, appID, tag.ID)
		if internalErr == nil {
			return errs.NewError(errs.InvalidOperation, fmt.Sprintf("tag %v already exists in app %v", value, appID))
		}

		if internalErr.Code != errs.NotFound {
			return internalErr
		}

		appTagRelation = entity.AppTagRelation{
			AppID: appID,
			TagID: tag.ID,
		}

		return a.appTagRelationDao.CreateAppTagRelation(ct, tx, appTagRelation)
	})

	return tag, err
}

func (a App) RemoveTagFromApp(ct context.Context, appID uint64, tagID uint64) (entity.Tag, *errs.Error) {
	var tag entity.Tag
	txCtx := transaction.NewTransactionsContext(
		a.logger,
		a.transactionFactory,
		a.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		tag, internalErr = a.tagDao.FindTagByIDWithTx(ct, tx, tagID)
		if internalErr != nil {
			return internalErr
		}

		return a.appTagRelationDao.DeleteAppTagRelationByAppIDAndTagID(ct, tx, appID, tagID)
	})

	return tag, err
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
	txCtx := transaction.NewTransactionsContext(
		a.logger,
		a.transactionFactory,
		a.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		app, internalErr = a.appDao.FindAppByIDWithTx(ct, tx, appID)
		if internalErr != nil {
			return internalErr
		}

		internalErr = a.appDao.DeleteApp(ct, tx, appID)
		if internalErr != nil {
			return internalErr
		}

		return nil
	})

	if err != nil {
		return entity.App{}, err
	}

	// TODO: delete registered app resource and groups

	return app, nil
}

func (a App) FindAppVersionsByAppID(ct context.Context, appID uint64) ([]entity.AppVersion, *errs.Error) {
	return a.appVersionDao.FindAppVersionsByAppID(ct, appID)
}

func (a App) CreateAppVersion(ct context.Context, appID uint64, createAppVersionInput CreateAppVersionInput) (entity.AppVersion, *errs.Error) {
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
		AppID:           appID,
		AppName:         createAppVersionInput.AppName,
		Description:     createAppVersionInput.Description,
		CreatedByUserID: userID,
		IsReady:         false,
		CreatedAt:       time.Now().UTC(),
	}
	txCtx := transaction.NewTransactionsContext(
		a.logger,
		a.transactionFactory,
		a.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		maxVersion, err := a.appVersionDao.FindMaxVersionNumberWithTx(ct, tx, appID)
		if err != nil {
			if err.Code == errs.NotFound {
				// no version exists, start from 0
				maxVersion = 0
			} else {
				return err
			}
		}

		av.Number = maxVersion + 1
		err = a.appVersionDao.CreateAppVersion(ct, tx, av)
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

func (a App) CreateAppPackageFileUploadSession(ct context.Context, appID uint64, versionNumber int) (uint64, *errs.Error) {
	currUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return 0, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if a.featureToggles.EnableAuthorization {
		query := authorization.NewUpdateAppVersionInAppQuery(currUserID, currUserID)
		hasPermission, err := a.authorizer.HasPermission(ct, query)
		if err != nil {
			return 0, err
		}

		if !hasPermission {
			return 0, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	res, rpcErr := a.cloudClientRegistry.FileClient().CreateUploadSession(ct, &emptypb.Empty{})
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return 0, internalErr
	}

	fileUploadSession := entity.AppPackageUploadSession{
		AppID:               appID,
		VersionNumber:       versionNumber,
		FileUploadSessionID: res.UploadSessionId,
		IsCompleted:         false,
		CreatedAt:           time.Now().UTC(),
	}
	txCtx := transaction.NewTransactionsContext(
		a.logger,
		a.transactionFactory,
		a.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		return a.appPackageUploadSessionDao.CreateAppPackageUploadSession(ct, tx, fileUploadSession)
	})
	return res.UploadSessionId, err
}

func (a App) FinishAppPackageFileUploadSession(ct context.Context, appID uint64, versionNumber int, fileUploadSessionID uint64) (entity.AppVersion, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.AppVersion{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if a.featureToggles.EnableAuthorization {
		query := authorization.NewUpdateAppVersionInAppQuery(userID, userID)
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

	findUploadSessionReq := proto.FindUploadSessionRequest{
		UploadSessionId: fileUploadSessionID,
	}
	uploadSession, rpcErr := a.cloudClientRegistry.FileClient().FindUploadSession(ct, &findUploadSessionReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.AppVersion{}, internalErr
	}
	var appVersion entity.AppVersion
	txCtx := transaction.NewTransactionsContext(
		a.logger,
		a.transactionFactory,
		a.stateSyncer,
		ct,
	)

	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		appPackageUploadSession, err := a.appPackageUploadSessionDao.FindAppPackageUploadSessionWithTx(
			ct,
			tx,
			appID,
			versionNumber,
			fileUploadSessionID)
		if err != nil {
			return err
		}

		if appPackageUploadSession.IsCompleted {
			return errs.NewError(errs.InvalidOperation, fmt.Sprintf("app package upload session is already succeeded: userID=%v, appID=%v, versionNumber=%v, fileUploadSessionID=%v",
				userID,
				appID,
				versionNumber,
				fileUploadSessionID))
		}

		now := time.Now().UTC()
		appPackageUploadSession.IsCompleted = true
		appPackageUploadSession.UpdatedAt = &now
		return a.appPackageUploadSessionDao.UpdateAppPackageFileUploadSession(ct, tx, appPackageUploadSession)
	})

	go func() {
		ct := context.Background()
		err := a.uploadAppPackageFiles(ct, userID, appID, versionNumber, uploadSession)
		if err != nil {
			a.logger.ErrorWithContext(ct, err)
			return
		}
		// TODO: change app version status to ready
	}()

	return appVersion, err
}

func (a App) DeleteAppVersion(ct context.Context, appID uint64, versionNumber int) (entity.AppVersion, *errs.Error) {
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
	txCtx := transaction.NewTransactionsContext(
		a.logger,
		a.transactionFactory,
		a.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		av, internalErr = a.appVersionDao.FindAppVersionByAppIDAndVersionNumberWithTx(ct, tx, appID, versionNumber)
		if internalErr != nil {
			return internalErr
		}

		if av.Locked {
			return errs.NewError(errs.InvalidOperation, fmt.Sprintf("app version %v is locked", versionNumber))
		}

		return a.appVersionDao.DeleteAppVersion(ct, tx, appID, versionNumber)
	})

	return av, err
}

func (a App) FindPricesByAppVersionID(ct context.Context, appID uint64, versionNumber int) ([]entity.Money, *errs.Error) {
	return a.appVersionPriceDao.FindAppVersionPricesByAppIDAndVersionNumber(ct, appID, versionNumber)
}

func (a App) FindAppVersionChangesByAppVersionID(ct context.Context, appID uint64, versionNumber int) ([]string, *errs.Error) {
	return a.appVersionChangeDao.FindAppVersionChangesByAppIDAndVersionNumber(ct, appID, versionNumber)
}

func (a App) GetAppSecretToken(ct context.Context, generateTokenInput GenerateTokenInput) (string, *errs.Error) {
	return a.jwtAuthority.GenerateToken(ct, generateTokenInput)
}

func (a App) uploadAppPackageFiles(
	ct context.Context,
	userID uint64,
	appID uint64,
	versionNumber int,
	uploadSession *proto.UploadSession,
) *errs.Error {
	fileReader, err := a.storageMapClient.Get(strconv.FormatInt(int64(uploadSession.FileId), 10))
	if err != nil {
		return err
	}

	buffedFiledReader := bufio.NewReaderSize(fileReader, bufferSize)
	gzipReader, error := gzip.NewReader(buffedFiledReader)
	if error != nil {
		return errs.NewError(errs.IO, error.Error())
	}

	tarReader := tar.NewReader(gzipReader)
	appIDStr := strconv.FormatInt(int64(appID), 10)
	versionNumberStr := strconv.Itoa(versionNumber)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return errs.NewError(errs.IO, err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			// TODO: use separate storage map client for reading gzip file and for uploading extracted files
			headerName := a.removeTarFilePrefix(header.Name)
			err := a.processFile(ct, userID, appID, versionNumber, tarReader, headerName, func(reader io.Reader) *errs.Error {
				storageMapKey := path.Join(appPackageRoot, appIDStr, versionNumberStr, headerName)
				return a.storageMapClient.Put(storageMapKey, reader)
			})

			if err != nil {
				return err
			}
		default:
			return errs.NewError(errs.IO, fmt.Sprintf("unknown type: %v in %s", header.Typeflag, header.Name))
		}
	}

	return nil
}

/*
removeTarFilePrefix removes the first directory name from file header name.
The file extracted from the tar will have the name of tar as prefix in the file name.
The prefix is defined by the user
Eg:

	tar name: app.tar
	before tar: manifest.yaml
	after extract from tar: app/manifest.yaml
*/
func (a App) removeTarFilePrefix(fileHeaderName string) string {
	parts := strings.Split(fileHeaderName, string(os.PathSeparator))
	return path.Join(parts[1:]...)
}

func (a App) processFile(ct context.Context, userID uint64, appID uint64, versionNumber int, reader *tar.Reader, fileName string, uploaderFunc UploadFunc) *errs.Error {
	switch fileName {
	case "manifest.yaml":
		return a.processManifestFile(ct, userID, appID, versionNumber, reader, uploaderFunc)
	default:
		return uploaderFunc(reader)
	}
}

func (a App) processManifestFile(ct context.Context, userID uint64, appID uint64, versionNumber int, reader *tar.Reader, uploaderFunc UploadFunc) *errs.Error {
	manifestData := struct {
		AppName        string         `yaml:"app_name"`
		Description    string         `yaml:"description"`
		HasUiExtension bool           `yaml:"has_ui_extension"`
		Changes        []string       `yaml:"changes"`
		Prices         map[string]int `yaml:"prices"`
	}{}
	readers := tmio.NewMultiReaders(reader, 2)
	extractReader, uploadReader := readers[0], readers[1]

	wg := sync.WaitGroup{}
	var wgErr *errs.Error
	once := sync.Once{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := yaml.NewDecoder(extractReader).Decode(&manifestData)
		if err != nil {
			once.Do(func() {
				wgErr = errs.NewError(errs.InvalidArgument, err.Error())
			})
		}
	}()

	go func() {
		defer wg.Done()
		err := uploaderFunc(uploadReader)
		if err != nil {
			once.Do(func() {
				wgErr = err
			})
		}
	}()

	wg.Wait()
	if wgErr != nil {
		return wgErr
	}

	appVersion := entity.AppVersion{
		AppID:           appID,
		Number:          versionNumber,
		CreatedByUserID: userID,
		AppName:         manifestData.AppName,
		Description:     manifestData.Description,
		HasUiExtension:  manifestData.HasUiExtension,
		IsReady:         true,
		CreatedAt:       time.Now().UTC(),
	}
	txCtx := transaction.NewTransactionsContext(
		a.logger,
		a.transactionFactory,
		a.stateSyncer,
		ct,
	)
	return txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		_, err := a.appVersionDao.FindAppVersionByAppIDAndVersionNumberWithTx(ct, tx, appID, versionNumber)
		if err != nil {
			return err
		}

		err = a.appVersionDao.UpdateAppVersion(ct, tx, appVersion)
		if err != nil {
			return err
		}

		for currency, price := range manifestData.Prices {
			appVersionPrice := entity.AppVersionPrice{
				Money: entity.Money{
					Currency: entity.Currency(currency),
					Amount:   price,
				},
				AppID:         appVersion.AppID,
				VersionNumber: appVersion.Number,
			}
			err := a.appVersionPriceDao.CreateAppVersionPrice(ct, tx, appVersionPrice)
			if err != nil {
				return err
			}
		}

		for _, change := range manifestData.Changes {
			genAppSecretIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "appVersionChangeID"}
			genAppSecretIDRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genAppSecretIDReq)
			if rpcErr != nil {
				return errs.FromGRPCErr(rpcErr)
			}

			change := entity.AppVersionChange{
				AppID:         appVersion.AppID,
				ChangeID:      genAppSecretIDRes.UniqueNumber,
				VersionNumber: appVersion.Number,
				Change:        change,
			}
			err := a.appVersionChangeDao.CreateAppVersionChange(ct, tx, change)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func NewApp(
	logger telemetry.Logger,
	storageMapClient storage.MapClient,
	cloudClientRegistry *client.Registry,
	authorizer client.Authorizer,
	featureToggles feature.Toggles,
	transactionFactory cloudTransaction.Factory,
	stateSyncer *realtime.StateSyncer,
	appDao dao.App,
	appVersionDao dao.AppVersion,
	appVersionPriceDao dao.AppVersionPrice,
	appVersionChangeDao dao.AppVersionChange,
	appSecretDao dao.AppSecret,
	appPackageUploadSessionDao dao.AppPackageUploadSession,
	teamAppInstallationDao dao.TeamAppInstallation,
	teamDao dao.Team,
	tagDao dao.Tag,
	appTagRelationDao dao.AppTagRelation,
	appGroupRelationDao dao.AppGroupRelation,
	groupRolloutRelationDao dao.GroupRolloutRelation,
	groupMemberRelationDao dao.GroupMemberRelation,
	appRolloutRelation dao.AppRolloutRelation,
	rolloutDao dao.Rollout,
	groupRepository *repository.Group,
	activatorRepository *repository.Activator,
	versionSelectorRepository *repository.VersionSelector,
	jwtAuthority security.JWTAuthority,
) App {
	return App{
		logger:                     logger,
		storageMapClient:           storageMapClient,
		cloudClientRegistry:        cloudClientRegistry,
		authorizer:                 authorizer,
		featureToggles:             featureToggles,
		transactionFactory:         transactionFactory,
		stateSyncer:                stateSyncer,
		appDao:                     appDao,
		appVersionDao:              appVersionDao,
		appVersionPriceDao:         appVersionPriceDao,
		appVersionChangeDao:        appVersionChangeDao,
		appSecretDao:               appSecretDao,
		appPackageUploadSessionDao: appPackageUploadSessionDao,
		teamAppInstallationDao:     teamAppInstallationDao,
		teamDao:                    teamDao,
		tagDao:                     tagDao,
		appTagRelationDao:          appTagRelationDao,
		appGroupRelationDao:        appGroupRelationDao,
		groupMemberRelationDao:     groupMemberRelationDao,
		groupRolloutRelationDao:    groupRolloutRelationDao,
		appRolloutRelation:         appRolloutRelation,
		rolloutDao:                 rolloutDao,
		groupRepository:            groupRepository,
		activatorRepository:        activatorRepository,
		versionSelectorRepository:  versionSelectorRepository,
		jwtAuthority:               jwtAuthority,
	}
}
