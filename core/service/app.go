package service

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"path"
	"strconv"
	"time"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/client"
	cloudAuthorization "github.com/teamyapp/cloud/libs/authorization"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/storage"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"google.golang.org/protobuf/types/known/emptypb"
)

var appPackageRoot = path.Join("app", "packages")

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
	transactionFactory         transaction.Factory
	stateSyncer                *realtime.StateSyncer
	appDao                     dao.App
	appVersionDao              dao.AppVersion
	appVersionPriceDao         dao.AppVersionPrice
	appVersionChangeDaO        dao.AppVersionChange
	appSecretDao               dao.AppSecret
	appPackageUploadSessionDao dao.AppPackageUploadSession
	teamAppInstallationDao     dao.TeamAppInstallation
	teamDao                    dao.Team
}

type AppFilter struct {
	AppID  *uint64
	TeamID *uint64
}

type UpdateAppTeamInstallationInput struct {
	EnabledVersionNumber int32
}

func (a App) FindAppByID(ct context.Context, appID uint64) (entity.App, *errs.Error) {
	return a.appDao.FindAppByID(ct, appID)
}

func (a App) FindAppSecretsByAppID(ct context.Context, appID uint64) ([]entity.AppSecret, *errs.Error) {
	return a.appSecretDao.FindAppSecretsByAppID(ct, appID)
}

func (a App) CreateAppSecret(ct context.Context, appID uint64, name string) (entity.AppSecret, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.AppSecret{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	genAppSecretIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "appSecretID"}
	genAppSecretIDRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genAppSecretIDReq)
	if rpcErr != nil {
		return entity.AppSecret{}, errs.FromGRPCErr(rpcErr)
	}

	appSecret := entity.AppSecret{
		ID:            genAppSecretIDRes.UniqueNumber,
		Name:          name,
		AppID:         appID,
		AddedAt:       time.Now().UTC(),
		AddedByUserID: userID,
	}
	return a.appSecretDao.CreateAppSecret(ct, appSecret)
}

func (a App) UpdateAppSecret(ct context.Context, appID uint64, appSecretID uint64, name string) (entity.AppSecret, *errs.Error) {
	var appSecret entity.AppSecret
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		appSecret, internalErr = a.appSecretDao.FindAppSecretByIDWithTx(ct, tx, appSecretID)
		if internalErr != nil {
			return internalErr
		}

		appSecret.Name = name
		return a.appSecretDao.UpdateAppSecretWithTx(ct, tx, appSecretID, appSecret)
	})

	if err != nil {
		return entity.AppSecret{}, err
	}

	return appSecret, nil
}

func (a App) DeleteAppSecret(ct context.Context, appSecretID uint64) (entity.AppSecret, *errs.Error) {
	var appSecret entity.AppSecret
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		appSecret, internalErr = a.appSecretDao.FindAppSecretByIDWithTx(ct, tx, appSecretID)
		if internalErr != nil {
			return internalErr
		}

		return a.appSecretDao.DeleteAppSecretWithTx(ct, tx, appSecretID)
	})

	if err != nil {
		return entity.AppSecret{}, err
	}

	return appSecret, nil
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

	return a.teamAppInstallationDao.CreateTeamAppInstallation(ct, teamAppInstallation)
}

func (a App) UninstallAppFromTeam(ct context.Context, appInstallationID uint64) (entity.TeamAppInstallation, *errs.Error) {
	var teamAppInstallation entity.TeamAppInstallation
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		teamAppInstallation, internalErr = a.teamAppInstallationDao.FindTeamAppInstallationByIDWithTx(ct, tx, appInstallationID)
		if internalErr != nil {
			return internalErr
		}

		return a.teamAppInstallationDao.DeleteTeamAppInstallationByIDWithTx(ct, tx, appInstallationID)
	})

	if err != nil {
		return entity.TeamAppInstallation{}, err
	}

	return teamAppInstallation, nil
}

func (a App) FindTeamAppInstallationsByAppID(ct context.Context, appID uint64) ([]entity.TeamAppInstallation, *errs.Error) {
	return a.teamAppInstallationDao.FindTeamAppInstallationsByAppID(ct, appID)
}

func (a App) FindAppVersionByAppIDAndNumber(ct context.Context, appID uint64, versionNumber int) (entity.AppVersion, *errs.Error) {
	return a.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
}

func (a App) CreateApp(ct context.Context, name string, teamID uint64) (entity.App, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.App{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	genAppIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "appID"}
	genAppIDRes, rpcErr := a.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genAppIDReq)
	if rpcErr != nil {
		return entity.App{}, errs.FromGRPCErr(rpcErr)
	}

	app := entity.App{
		ID:                 genAppIDRes.UniqueNumber,
		TotalInstallations: 0,
		ManagedByTeamID:    teamID,
		CreatedAt:          time.Now().UTC(),
	}
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		return a.appDao.CreateApp(ct, tx, app)
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

func (a App) CreateAppVersion(ct context.Context, appID uint64, appName string, description string) (entity.AppVersion, *errs.Error) {
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
		AppName:         appName,
		Description:     description,
		CreatedByUserID: userID,
		IsReady:         false,
		CreatedAt:       time.Now().UTC(),
	}
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}

	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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
		err := a.uploadAppPackageFiles(ct, appID, versionNumber, uploadSession)
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
	txCtx := TransactionsContext{
		logger:             a.logger,
		transactionFactory: a.transactionFactory,
		stateSyncer:        a.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		av, internalErr = a.appVersionDao.FindAppVersionByAppIDAndVersionNumberWithTx(ct, tx, appID, versionNumber)
		if internalErr != nil {
			return internalErr
		}

		return a.appVersionDao.DeleteAppVersion(ct, tx, appID, versionNumber)
	})

	if err != nil {
		return entity.AppVersion{}, err
	}

	return av, nil
}

func (a App) FindAppVersionPricesByAppVersionID(ct context.Context, appID uint64, versionNumber int) ([]entity.Money, *errs.Error) {
	return a.appVersionPriceDao.FindAppVersionPricesByAppIDAndVersionNumber(ct, appID, versionNumber)
}

func (a App) FindAppVersionChangesByAppVersionID(ct context.Context, appID uint64, versionNumber int) ([]string, *errs.Error) {
	return a.appVersionChangeDaO.FindAppVersionChangesByAppIDAndVersionNumber(ct, appID, versionNumber)
}

func (a App) uploadAppPackageFiles(
	ct context.Context,
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
			fullPath := path.Join(appPackageRoot, appIDStr, versionNumberStr, header.Name)
			err := a.storageMapClient.Put(fullPath, tarReader)
			if err != nil {
				return err
			}
		default:
			return errs.NewError(errs.IO, fmt.Sprintf("unknown type: %v in %s", header.Typeflag, header.Name))
		}
	}

	return nil
}

func NewApp(
	logger telemetry.Logger,
	storageMapClient storage.MapClient,
	cloudClientRegistry *client.Registry,
	authorizer client.Authorizer,
	featureToggles feature.Toggles,
	transactionFactory transaction.Factory,
	stateSyncer *realtime.StateSyncer,
	appDao dao.App,
	appVersionDao dao.AppVersion,
	appVersionPriceDao dao.AppVersionPrice,
	appVersionChangeDaO dao.AppVersionChange,
	appSecretDao dao.AppSecret,
	appPackageUploadSessionDao dao.AppPackageUploadSession,
	teamAppInstallationDao dao.TeamAppInstallation,
	teamDao dao.Team,
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
		appVersionChangeDaO:        appVersionChangeDaO,
		appSecretDao:               appSecretDao,
		appPackageUploadSessionDao: appPackageUploadSessionDao,
		teamAppInstallationDao:     teamAppInstallationDao,
		teamDao:                    teamDao,
	}
}
