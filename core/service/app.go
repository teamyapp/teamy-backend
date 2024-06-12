package service

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/teamyapp/cloud/app/client"
	cloudAuthorization "github.com/teamyapp/cloud/libs/authorization"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	tmio "github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/cloud/libs/randgen"
	"github.com/teamyapp/cloud/libs/security"
	"github.com/teamyapp/cloud/libs/storage"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudTransaction "github.com/teamyapp/cloud/libs/transaction"
	pbcloud "github.com/teamyapp/protocol/pb/pbgo/cloud"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/cache"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/repository"
	"github.com/teamyapp/teamy-backend/core/transaction"
	"google.golang.org/protobuf/types/known/emptypb"
	"gopkg.in/yaml.v3"
)

var appPackageRoot = path.Join("files", "apps")
var secretAlphabet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!?@#_-")
var secretLength = 32
var defaultAppOwnersGroupName = "App Owners"
var defaultPublicGroupName = "Public"
var defaultRolloutName = "default rollout"
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
	transactionGroupFactory    transaction.GroupFactory
	objectStore                storage.ObjectStore
	cloudClientRegistry        *client.Registry
	authorizer                 client.Authorizer
	featureToggles             feature.Toggles
	transactionFactory         cloudTransaction.Factory
	stateSyncer                *realtime.StateSyncer
	cache                      *cache.TimeBasedCache[string, any]
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
	appRolloutRelation         dao.AppRolloutRelation
	rolloutDao                 dao.Rollout
	groupRepository            *repository.Group
	activatorRepository        *repository.Activator
	versionSelectorRepository  *repository.VersionSelector
	jwtAuthority               security.JWTAuthority
	groupService               *Group
	rolloutService             *Rollout
}

type AppFilter struct {
	TagValues []string
}

type UpdateAppTeamInstallationInput struct {
	EnabledVersionNumber int32
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

type UpdateAppVersionInput struct {
	Status entity.AppVersionStatus
}

func (a App) FindApps(ct context.Context, appFilter *AppFilter) ([]entity.App, *errs.Error) {
	var apps []entity.App
	transactionErr := a.transactionGroupFactory.WithTransactionGroup(ct, true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		if appFilter == nil || len(appFilter.TagValues) == 0 {
			apps, err = a.appDao.FindAppsWithTx(ct, tx)
			return err
		} else {
			appIDs, err := a.appTagRelationDao.FindAppIDsByTagValuesWithTx(ct, tx, appFilter.TagValues)
			if err != nil {
				return err
			}

			apps, err = a.appDao.FindAppsByAppIDsWithTx(ct, tx, appIDs)
			return err
		}
	})
	return apps, transactionErr
}

func (a App) FindAppByID(ct context.Context, appID uint64) (entity.App, *errs.Error) {
	if a.featureToggles.EnableCache {
		value, cacheErr := a.cache.Get(ct, findAppByIDCacheKey(appID))
		if cacheErr == nil {
			return value.(entity.App), nil
		}

		var cacheKeyNotFoundErr cache.KeyNotFoundErr[string]
		if !errors.As(cacheErr, &cacheKeyNotFoundErr) {
			return entity.App{}, errs.NewError(errs.Unknown, cacheErr.Error())
		}
	}

	app, err := a.appDao.FindAppByID(ct, appID)
	if err != nil {
		return entity.App{}, err
	}

	if a.featureToggles.EnableCache {
		cacheErr := a.cache.SetIfExpired(ct, findAppByIDCacheKey(appID), app)
		if cacheErr != nil {
			return entity.App{}, errs.NewError(errs.Unknown, cacheErr.Error())
		}
	}

	return app, nil
}

func (a App) FindAppsByGroupID(ct context.Context, groupID uint64) ([]entity.App, *errs.Error) {
	var apps []entity.App
	err := a.transactionGroupFactory.WithTransactionGroup(ct, true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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

	genAppSecretIDReq := &pbcloud.GenerateUniqueNumberRequest{SequenceName: "appSecretID"}
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
	err = a.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		return a.appSecretDao.CreateAppSecret(ct, tx, appSecret)
	})
	return appSecret, err
}

func (a App) UpdateAppSecret(ct context.Context, appSecretID uint64, input UpdateAppSecretInput) (entity.AppSecret, *errs.Error) {
	var appSecret entity.AppSecret
	err := a.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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
	err := a.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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
	genTeamAppInstallationIDReq := &pbcloud.GenerateUniqueNumberRequest{SequenceName: "teamAppInstallationID"}
	genTeamAppInstallationIDRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genTeamAppInstallationIDReq)
	if rpcErr != nil {
		return entity.TeamAppInstallation{}, errs.FromGRPCErr(rpcErr)
	}

	teamAppInstallation := entity.TeamAppInstallation{
		ID:              genTeamAppInstallationIDRes.UniqueNumber,
		AppID:           appID,
		InstalledTeamID: teamID,
	}
	err := a.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		err := a.teamAppInstallationDao.CreateTeamAppInstallation(ct, tx, teamAppInstallation)
		if err != nil {
			return err
		}

		app, err := a.appDao.FindAppByID(ct, appID)
		if err != nil {
			return err
		}

		app.TotalInstallations++
		now := time.Now().UTC()
		app.UpdatedAt = &now
		return a.appDao.UpdateApp(ct, tx, app)
	})
	return teamAppInstallation, err
}

func (a App) UninstallAppFromTeam(ct context.Context, appInstallationID uint64) (entity.TeamAppInstallation, *errs.Error) {
	var teamAppInstallation entity.TeamAppInstallation
	err := a.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		teamAppInstallation, internalErr = a.teamAppInstallationDao.FindTeamAppInstallationByIDWithTx(ct, tx, appInstallationID)
		if internalErr != nil {
			return internalErr
		}

		err := a.teamAppInstallationDao.DeleteTeamAppInstallationByID(ct, tx, appInstallationID)
		if err != nil {
			return err
		}

		app, err := a.appDao.FindAppByID(ct, teamAppInstallation.AppID)
		if err != nil {
			return err
		}

		app.TotalInstallations--
		now := time.Now().UTC()
		app.UpdatedAt = &now
		return a.appDao.UpdateApp(ct, tx, app)
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
	if a.featureToggles.EnableCache {
		value, cacheErr := a.cache.Get(ct, findAppVersionByAppIDAndNumberCacheKey(appID, versionNumber))
		if cacheErr == nil {
			return value.(entity.AppVersion), nil
		}

		var cacheKeyNotFoundErr cache.KeyNotFoundErr[string]
		if !errors.As(cacheErr, &cacheKeyNotFoundErr) {
			return entity.AppVersion{}, errs.NewError(errs.Unknown, cacheErr.Error())
		}
	}

	appVersion, err := a.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
	if err != nil {
		return entity.AppVersion{}, err
	}

	if a.featureToggles.EnableCache {
		cacheErr := a.cache.SetIfExpired(ct, findAppVersionByAppIDAndNumberCacheKey(appID, versionNumber), appVersion)
		if cacheErr != nil {
			return entity.AppVersion{}, errs.NewError(errs.Unknown, cacheErr.Error())
		}
	}

	return appVersion, nil
}

func (a App) CreateApp(ct context.Context, teamID uint64, createAppInput CreateAppInput) (entity.App, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.App{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	genAppIDReq := &pbcloud.GenerateUniqueNumberRequest{SequenceName: "appID"}
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
		Status:          entity.AppVersionStatusReady,
		Locked:          true,
		CreatedAt:       now,
	}

	genGroupIDReq := &pbcloud.GenerateUniqueNumberRequest{SequenceName: "groupID"}
	genGroupIDRes, rpcErr := a.cloudClientRegistry.
		GeneratorClient().
		GenerateUniqueNumber(ct, genGroupIDReq)
	if rpcErr != nil {
		return entity.App{}, errs.FromGRPCErr(rpcErr)
	}

	appOwnersGroup := entity.StaticGroup{
		Group: entity.Group{
			ID:              genGroupIDRes.UniqueNumber,
			Type:            entity.GroupTypeStatic,
			MemberType:      entity.GroupMemberTypeTeam,
			Name:            defaultAppOwnersGroupName,
			MaxRolloutIndex: 0,
			CreatedAt:       now,
			Locked:          true,
		},
		MemberIDs: []uint64{teamID},
	}

	genGroupIDRes, rpcErr = a.cloudClientRegistry.
		GeneratorClient().
		GenerateUniqueNumber(ct, genGroupIDReq)
	if rpcErr != nil {
		return entity.App{}, errs.FromGRPCErr(rpcErr)
	}

	publicGroup := entity.FilterGroup{
		Group: entity.Group{
			ID:              genGroupIDRes.UniqueNumber,
			Type:            entity.GroupTypeFilter,
			MemberType:      entity.GroupMemberTypeTeam,
			Name:            defaultPublicGroupName,
			MaxRolloutIndex: 0,
			CreatedAt:       now,
			Locked:          true,
		},
		Filter: "true;",
	}

	genActivatorIDReq := &pbcloud.GenerateUniqueNumberRequest{SequenceName: "activatorID"}
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

	genVersionSelectorIDReq := &pbcloud.GenerateUniqueNumberRequest{SequenceName: "versionSelectorID"}
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

	genRolloutIDReq := &pbcloud.GenerateUniqueNumberRequest{SequenceName: "rolloutID"}
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
	internalErr := a.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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

		err = a.groupRepository.CreateStaticGroup(ct, tx, appOwnersGroup)
		if err != nil {
			return err
		}

		err = a.groupRepository.CreateFilterGroup(ct, tx, publicGroup)
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
			GroupID:    appOwnersGroup.ID,
			RolloutID:  rollout.ID,
			OrderIndex: 0,
		})
		if err != nil {
			return err
		}

		err = a.appGroupRelationDao.CreateAppGroupRelation(ct, tx, entity.AppGroupRelation{
			AppID:   app.ID,
			GroupID: appOwnersGroup.ID,
		})
		if err != nil {
			return err
		}

		err = a.appGroupRelationDao.CreateAppGroupRelation(ct, tx, entity.AppGroupRelation{
			AppID:   app.ID,
			GroupID: publicGroup.ID,
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
	var tags []entity.Tag
	err := a.transactionGroupFactory.WithTransactionGroup(ct, true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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
	err := a.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		tag, internalErr = a.tagDao.FindTagByValueWithTx(ct, tx, value)
		if internalErr != nil {
			if internalErr.Code != errs.NotFound {
				return internalErr
			}

			genTagIDReq := &pbcloud.GenerateUniqueNumberRequest{SequenceName: "tagID"}
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
	err := a.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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
	transactionErr := a.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		app, err = a.appDao.FindAppByIDWithTx(ct, tx, appID)
		if err != nil {
			return err
		}

		groupIDs, err := a.appGroupRelationDao.FindGroupIDsByAppIDWithTx(ct, tx, appID)
		if err != nil {
			return err
		}

		for _, groupID := range groupIDs {
			_, err = a.groupService.deleteGroup(ct, tx, groupID, true)
			if err != nil {
				return err
			}
		}

		versions, err := a.appVersionDao.FindAppVersionsByAppIDWithTx(ct, tx, appID)
		if err != nil {
			return err
		}

		for _, version := range versions {
			err = a.appVersionDao.DeleteAppVersion(ct, tx, appID, version.Number)
			if err != nil {
				return err
			}
		}

		tagIDs, err := a.appTagRelationDao.FindTagIDsByAppIDWithTx(ct, tx, appID)
		if err != nil {
			return err
		}

		for _, tagID := range tagIDs {
			err = a.appTagRelationDao.DeleteAppTagRelationByAppIDAndTagID(ct, tx, appID, tagID)
			if err != nil {
				return err
			}

			err = a.tagDao.DeleteTag(ct, tx, tagID)
			if err != nil {
				return err
			}
		}

		err = a.appSecretDao.DeleteAppSecretsByAppID(ct, tx, appID)
		if err != nil {
			return err
		}

		err = a.teamAppInstallationDao.DeleteTeamAppInstallationsByAppID(ct, tx, appID)
		if err != nil {
			return err
		}

		rolloutIDs, err := a.appRolloutRelation.FindRolloutIDsByAppIDWithTx(ct, tx, appID)
		if err != nil {
			return err
		}

		for _, rolloutID := range rolloutIDs {
			_, err = a.rolloutService.deleteRollout(ct, tx, rolloutID, true)
			if err != nil {
				return err
			}
		}

		err = a.appPackageUploadSessionDao.DeleteAppPackageUploadSessionsByAppID(ct, tx, appID)
		if err != nil {
			return err
		}

		return a.appDao.DeleteApp(ct, tx, appID)
	})

	return app, transactionErr
}

func (a App) FindAppVersionsByAppID(ct context.Context, appID uint64) ([]entity.AppVersion, *errs.Error) {
	return a.appVersionDao.FindAppVersionsByAppID(ct, appID)
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
		AppID:           appID,
		CreatedByUserID: userID,
		Status:          entity.AppVersionStatusInit,
		CreatedAt:       time.Now().UTC(),
		Locked:          false,
	}
	err := a.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		maxVersion, err := a.appVersionDao.FindMaxVersionNumberWithTx(ct, tx, appID)
		if err != nil {
			if err.Code == errs.NotFound {
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

	return av, err
}

func (a App) UpdateAppVersion(ct context.Context, appID uint64, versionNumber int, input UpdateAppVersionInput) (entity.AppVersion, *errs.Error) {
	var av entity.AppVersion
	err := a.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		av, internalErr = a.appVersionDao.FindAppVersionByAppIDAndVersionNumberWithTx(ct, tx, appID, versionNumber)
		if internalErr != nil {
			return internalErr
		}

		av.Status = input.Status
		if input.Status == entity.AppVersionStatusInit {
			av.ErrorMessage = nil
		}

		now := time.Now().UTC()
		av.UpdatedAt = &now
		return a.appVersionDao.UpdateAppVersion(ct, tx, av)
	})

	return av, err
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
	err := a.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		err := a.appPackageUploadSessionDao.CreateAppPackageUploadSession(ct, tx, fileUploadSession)
		if err != nil {
			return err
		}

		appVersion, err := a.appVersionDao.FindAppVersionByAppIDAndVersionNumberWithTx(ct, tx, appID, versionNumber)
		if err != nil {
			return err
		}

		if appVersion.Status != entity.AppVersionStatusInit &&
			appVersion.Status != entity.AppVersionStatusError {
			return errs.NewError(errs.InvalidOperation, fmt.Sprintf("app version is not in init status: appID=%v, versionNumber=%v", appID, versionNumber))
		}

		appVersion.Status = entity.AppVersionStatusUploading
		now := time.Now().UTC()
		appVersion.UpdatedAt = &now

		updateAppVersionMutation := mutation.NewUpdateAppVersion(
			a.logger,
			a.stateSyncer,
			a.appVersionDao,
			a.appDao,
			appVersion,
		)

		err = updateAppVersionMutation.Execute(ct, tx)
		if err != nil {
			return err
		}

		rtTx.AppendMutation(updateAppVersionMutation)
		return nil
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

	findUploadSessionReq := pbcloud.FindUploadSessionRequest{
		UploadSessionId: fileUploadSessionID,
	}
	uploadSession, rpcErr := a.cloudClientRegistry.FileClient().FindUploadSession(ct, &findUploadSessionReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.AppVersion{}, internalErr
	}
	var appVersion entity.AppVersion
	err := a.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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
		err = a.appPackageUploadSessionDao.UpdateAppPackageFileUploadSession(ct, tx, appPackageUploadSession)
		if err != nil {
			return err
		}

		appVersion, err = a.appVersionDao.FindAppVersionByAppIDAndVersionNumberWithTx(ct, tx, appID, versionNumber)
		if err != nil {
			return err
		}

		appVersion.Status = entity.AppVersionStatusProcessing
		appVersion.UpdatedAt = &now
		updateAppVersionMutation := mutation.NewUpdateAppVersion(
			a.logger,
			a.stateSyncer,
			a.appVersionDao,
			a.appDao,
			appVersion,
		)

		err = updateAppVersionMutation.Execute(ct, tx)
		if err != nil {
			return err
		}

		rtTx.AppendMutation(updateAppVersionMutation)
		return nil
	})

	go func() {
		bgCt := context.Background()
		err = a.uploadAppPackageFiles(bgCt, userID, appID, versionNumber, uploadSession)
		if err != nil {
			a.logger.ErrorWithContext(bgCt, err)
			errMessage := err.Message
			transactionErr := a.transactionGroupFactory.WithTransactionGroup(bgCt, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
				appVersion, err = a.appVersionDao.FindAppVersionByAppIDAndVersionNumberWithTx(bgCt, tx, appID, versionNumber)
				if err != nil {
					return err
				}

				now := time.Now().UTC()
				appVersion.Status = entity.AppVersionStatusError
				appVersion.UpdatedAt = &now
				appVersion.ErrorMessage = &errMessage
				updateAppVersionMutation := mutation.NewUpdateAppVersion(
					a.logger,
					a.stateSyncer,
					a.appVersionDao,
					a.appDao,
					appVersion,
				)

				err = updateAppVersionMutation.Execute(bgCt, tx)
				if err != nil {
					return err
				}

				rtTx.AppendMutation(updateAppVersionMutation)
				return nil
			})

			if transactionErr != nil {
				a.logger.ErrorWithContext(bgCt, transactionErr)
			}

			return
		}
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
	err := a.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error

		rolloutIDs, internalErr := a.appRolloutRelation.FindRolloutIDsByAppIDWithTx(ct, tx, appID)
		if internalErr != nil {
			return internalErr
		}

		for _, rolloutID := range rolloutIDs {
			rollout, internalErr := a.rolloutDao.FindRolloutByIDWithTx(ct, tx, rolloutID)
			if internalErr != nil {
				return internalErr
			}

			versionSelector, internalErr := a.versionSelectorRepository.FindVersionSelectorByID(ct, tx, rollout.SelectorID)
			if internalErr != nil {
				return internalErr
			}

			switch versionSelector.Type {
			case entity.VersionSelectorTypeStatic:
				if versionSelector.StaticVersionSelector.VersionNumber == versionNumber {
					return errs.NewError(errs.InvalidOperation, fmt.Sprintf("version %v is used in rollout %v", versionNumber, rolloutID))
				}
			case entity.VersionSelectorTypeExperiment:
				versionNumbers := collect.Filter(versionSelector.ExperimentVersionSelector.VersionNumbers, func(vn int) bool {
					return vn == versionNumber
				})

				if len(versionNumbers) > 0 {
					return errs.NewError(errs.InvalidOperation, fmt.Sprintf("version %v is used in rollout %v", versionNumber, rolloutID))
				}
			default:
				return errs.NewError(errs.InvalidOperation, fmt.Sprintf("unknown version selector type: %v", versionSelector.Type))
			}
		}

		av, internalErr = a.appVersionDao.FindAppVersionByAppIDAndVersionNumberWithTx(ct, tx, appID, versionNumber)
		if internalErr != nil {
			return internalErr
		}

		if av.Locked {
			return errs.NewError(errs.InvalidOperation, fmt.Sprintf("app version %v is locked", versionNumber))
		}

		err := a.appVersionPriceDao.DeleteAppVersionPrice(ct, tx, appID, versionNumber)
		if err != nil {
			return err
		}

		return a.appVersionDao.DeleteAppVersion(ct, tx, appID, versionNumber)
	})

	return av, err
}

func (a App) FindPricesByAppVersionID(ct context.Context, appID uint64, versionNumber int) ([]entity.Money, *errs.Error) {
	return a.appVersionPriceDao.FindAppVersionPricesByAppIDAndVersionNumber(ct, appID, versionNumber)
}

func (a App) FindAppVersionChangesByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int) ([]entity.AppVersionChange, *errs.Error) {
	if a.featureToggles.EnableCache {
		cachedChanges, cacheErr := a.cache.Get(ct, findAppVersionChangesByAppIDAndVersionNumberCacheKey(appID, versionNumber))
		if cacheErr == nil {
			return cachedChanges.([]entity.AppVersionChange), nil
		}

		var cacheKeyNotFoundErr cache.KeyNotFoundErr[string]
		if !errors.As(cacheErr, &cacheKeyNotFoundErr) {
			return nil, errs.NewError(errs.Unknown, cacheErr.Error())
		}
	}

	appVersionChanges, err := a.appVersionChangeDao.FindAppVersionChangesByAppIDAndVersionNumber(ct, appID, versionNumber)
	if err != nil {
		return nil, err
	}

	if a.featureToggles.EnableCache {
		cacheErr := a.cache.SetIfExpired(ct, findAppVersionChangesByAppIDAndVersionNumberCacheKey(appID, versionNumber), appVersionChanges)
		if cacheErr != nil {
			return nil, errs.NewError(errs.Unknown, cacheErr.Error())
		}
	}

	return appVersionChanges, nil
}

func (a App) GetAppSecretToken(ct context.Context, generateTokenInput GenerateTokenInput) (string, *errs.Error) {
	return a.jwtAuthority.GenerateToken(ct, generateTokenInput)
}

func (a App) uploadAppPackageFiles(
	ct context.Context,
	userID uint64,
	appID uint64,
	versionNumber int,
	uploadSession *pbcloud.UploadSession,
) *errs.Error {
	fileReader, internalErr := a.objectStore.Get(ct, strconv.FormatInt(int64(uploadSession.FileId), 10))
	if internalErr != nil {
		return internalErr
	}

	buffedFiledReader := bufio.NewReaderSize(fileReader, bufferSize)
	gzipReader, err := gzip.NewReader(buffedFiledReader)
	if err != nil {
		return errs.NewError(errs.IO, err.Error())
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
				ext := filepath.Ext(headerName)
				mimeType := mime.TypeByExtension(ext)
				return a.objectStore.Put(ct, storageMapKey, reader, storage.ObjectMetadataInput{
					ContentType: mimeType,
				})
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
		AppName        string         `yaml:"appName"`
		Description    string         `yaml:"description"`
		HasUiExtension bool           `yaml:"hasUiExtension"`
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

	return a.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		appVersion, err := a.appVersionDao.FindAppVersionByAppIDAndVersionNumberWithTx(ct, tx, appID, versionNumber)
		if err != nil {
			return err
		}

		for currencyStr, price := range manifestData.Prices {
			currency, ok := entity.StringToCurrency[currencyStr]
			if !ok {
				return errs.NewError(errs.InvalidArgument, fmt.Sprintf("invalid currency: %v", currencyStr))
			}

			appVersionPrice := entity.AppVersionPrice{
				Money: entity.Money{
					Currency: currency,
					Amount:   price,
				},
				AppID:         appVersion.AppID,
				VersionNumber: appVersion.Number,
			}
			err = a.appVersionPriceDao.CreateAppVersionPrice(ct, tx, appVersionPrice)
			if err != nil {
				return err
			}
		}

		for _, change := range manifestData.Changes {
			genAppSecretIDReq := &pbcloud.GenerateUniqueNumberRequest{SequenceName: "appVersionChangeID"}
			genAppSecretIDRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genAppSecretIDReq)
			if rpcErr != nil {
				return errs.FromGRPCErr(rpcErr)
			}

			newChange := entity.AppVersionChange{
				ID:            genAppSecretIDRes.UniqueNumber,
				AppID:         appVersion.AppID,
				VersionNumber: appVersion.Number,
				Change:        change,
			}

			createAppVersionChangeMutation := mutation.NewCreateAppVersionChange(
				a.logger,
				a.stateSyncer,
				newChange,
				a.appVersionChangeDao,
				a.appDao,
			)

			err = createAppVersionChangeMutation.Execute(ct, tx)
			if err != nil {
				return err
			}

			rtTx.AppendMutation(createAppVersionChangeMutation)
		}

		now := time.Now().UTC()
		appVersion.Number = versionNumber
		appVersion.AppName = manifestData.AppName
		appVersion.Description = manifestData.Description
		appVersion.HasUiExtension = manifestData.HasUiExtension
		appVersion.Status = entity.AppVersionStatusReady
		appVersion.UpdatedAt = &now
		updateAppVersionMutation := mutation.NewUpdateAppVersion(
			a.logger,
			a.stateSyncer,
			a.appVersionDao,
			a.appDao,
			appVersion,
		)

		err = updateAppVersionMutation.Execute(ct, tx)
		if err != nil {
			return err
		}

		rtTx.AppendMutation(updateAppVersionMutation)
		return nil
	})
}

func NewApp(
	logger telemetry.Logger,
	transactionGroupFactory transaction.GroupFactory,
	objectStore storage.ObjectStore,
	cloudClientRegistry *client.Registry,
	authorizer client.Authorizer,
	featureToggles feature.Toggles,
	transactionFactory cloudTransaction.Factory,
	stateSyncer *realtime.StateSyncer,
	cache *cache.TimeBasedCache[string, any],
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
	appRolloutRelation dao.AppRolloutRelation,
	rolloutDao dao.Rollout,
	groupRepository *repository.Group,
	activatorRepository *repository.Activator,
	versionSelectorRepository *repository.VersionSelector,
	jwtAuthority security.JWTAuthority,
	groupService *Group,
	rolloutService *Rollout,
) App {
	return App{
		logger:                     logger,
		transactionGroupFactory:    transactionGroupFactory,
		objectStore:                objectStore,
		cloudClientRegistry:        cloudClientRegistry,
		authorizer:                 authorizer,
		featureToggles:             featureToggles,
		transactionFactory:         transactionFactory,
		stateSyncer:                stateSyncer,
		cache:                      cache,
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
		groupRolloutRelationDao:    groupRolloutRelationDao,
		appRolloutRelation:         appRolloutRelation,
		rolloutDao:                 rolloutDao,
		groupRepository:            groupRepository,
		activatorRepository:        activatorRepository,
		versionSelectorRepository:  versionSelectorRepository,
		jwtAuthority:               jwtAuthority,
		groupService:               groupService,
		rolloutService:             rolloutService,
	}
}

func findAppVersionByAppIDAndNumberCacheKey(appID uint64, versionNumber int) string {
	return fmt.Sprintf("FindAppVersionByAppIDAndNumber(%v,%v)", appID, versionNumber)
}

func findAppByIDCacheKey(appID uint64) string {
	return fmt.Sprintf("FindAppByID(%v)", appID)
}

func findAppVersionChangesByAppIDAndVersionNumberCacheKey(appID uint64, versionNumber int) string {
	return fmt.Sprintf("FindAppVersionChangesByAppIDAndVersionNumber(%v,%v)", appID, versionNumber)
}
