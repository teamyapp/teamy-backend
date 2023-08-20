package service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	cloudClient "github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/metrics/metricstest"
	"github.com/teamyapp/cloud/libs/network/networktest"
	"github.com/teamyapp/cloud/libs/retry"
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/rpc"
	"github.com/teamyapp/cloud/libs/runtime"
	"github.com/teamyapp/cloud/libs/storage"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/cloud/testkit"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/dao/daotest"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/service/servicetest"
)

type AppTestRef struct {
	appService               App
	appDao                   dao.App
	appVersionDao            dao.AppVersion
	appVersionVisibleTeamDao dao.AppVersionVisibleTeam
	teamDao                  dao.Team
	appTeamInstallationDao   dao.AppTeamInstallation
	transactionFactory       transaction.Factory
	cloudTestKit             testkit.TestKit
}

func prepareAppTestRef(t *testing.T, toggles feature.Toggles) (AppTestRef, bool) {
	lineFormatter := telemetry.NewOrderedColumnLineFormatter([]string{})
	logger := telemetry.NewLogger(lineFormatter, os.Stdout, telemetry.Off, []telemetry.LogInterceptor{})
	virtualNetwork := networktest.NewVirtualNetwork()
	cloudTestKitConfig := testkit.Config{
		GenUniqueNumberRangeSize: 10,
		JWTSigningKey:            "key",
		AccessTokenTTL:           2 * time.Hour,
		WebAPIBaseURL:            fmt.Sprintf("http://%s:%d", testkit.WebServerHost, testkit.WebServerPort),
		GithubClientID:           "123",
		GithubClientSecret:       "GithubSecret",
		GoogleClientID:           "456",
		GoogleClientSecret:       "GoogleSecret",
		SlackClientID:            "789",
		SlackClientSecret:        "SlackSecret",
		WebServerPort:            80,
		GRPCServerPort:           81,
	}
	cloudTestKit, internalErr := testkit.New(cloudTestKitConfig, virtualNetwork)
	require.Nil(t, internalErr)

	ct := context.Background()
	var accountOwner uint64 = 0
	serviceAccountID, internalErr := cloudTestKit.IdentityService.CreateServiceAccount(ct, accountOwner, "test")
	require.Nil(t, internalErr)

	apiToken, internalErr := cloudTestKit.IdentityService.GenerateServiceToken(ct, accountOwner, serviceAccountID)
	require.Nil(t, internalErr)

	testkit.StartServiceInstance(cloudTestKitConfig, virtualNetwork, cloudTestKit.ServiceInstanceRunner)

	teamyPrometheus := metricstest.NewNoopMetrics()
	cloudClientCfg := rpc.ConnectionConfig{
		Host:          testkit.GRPCServerHost,
		Port:          testkit.GRPCServerPort,
		ShouldEncrypt: false,
		GetAccessToken: func() string {
			return apiToken
		},
		RequestTimeout: 10 * time.Second,
	}
	cloudClientRegistry, err := cloudClient.NewRegistry(
		logger,
		virtualNetwork,
		teamyPrometheus,
		cloudClientCfg,
		func() retry.Retry {
			exponentialBackOff := backoff.NewExponentialBuilder().Build()
			return retry.NewMaxCount(
				logger,
				runtime.NewBuiltInRuntime(),
				exponentialBackOff,
				exponentialBackOff,
				3,
				nil)
		})
	require.Nil(t, err)

	authorizer := cloudClient.NewAuthorizer(logger, cloudClientRegistry)

	transactionFactory := transaction.NewFactory(nil)

	teamyBackendDB := dbtest.NewInMemoryDB()
	teamyBackendDB.CreateTable(daotest.AppTableName)
	teamyBackendDB.CreateTable(daotest.AppVersionTableName)
	teamyBackendDB.CreateTable(daotest.AppVersionVisibleTeamTableName)
	teamyBackendDB.CreateTable(daotest.AppTeamInstallationTableName)
	teamyBackendDB.CreateTable(daotest.TeamTableName)
	teamyBackendDB.CreateTable(daotest.AppPackageUploadSessionTableName)

	teamMemberDao := daotest.NewTeamMember(teamyBackendDB, transactionFactory)
	stateSyncer := realtime.NewStateSyncer(logger, teamMemberDao)

	appDao := daotest.NewApp(teamyBackendDB, transactionFactory)
	appPackageUploadSession := daotest.NewAppPackageUploadSession(teamyBackendDB)
	appVersionDao := daotest.NewAppVersion(teamyBackendDB, transactionFactory)
	appVersionVisibleTeamDao := daotest.NewAppVersionVisibleTeam(teamyBackendDB, transactionFactory)
	appTeamInstallationDao := daotest.NewAppTeamInstallation(teamyBackendDB, transactionFactory)
	teamDao := daotest.NewTeam(teamyBackendDB, transactionFactory)
	storageMapClient := storage.NewHTTPClient(cloudTestKitConfig.WebAPIBaseURL)

	appService := NewApp(
		logger,
		storageMapClient,
		cloudClientRegistry,
		authorizer,
		toggles,
		transactionFactory,
		stateSyncer,
		appDao,
		appVersionDao,
		appTeamInstallationDao,
		appVersionVisibleTeamDao,
		appPackageUploadSession,
		teamDao,
	)

	return AppTestRef{
		appService:               appService,
		transactionFactory:       transactionFactory,
		appDao:                   appDao,
		appVersionDao:            appVersionDao,
		appVersionVisibleTeamDao: appVersionVisibleTeamDao,
		teamDao:                  teamDao,
		appTeamInstallationDao:   appTeamInstallationDao,
		cloudTestKit:             cloudTestKit,
	}, true
}

func TestAppService_CreateApp(t *testing.T) {
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(appTestRef AppTestRef, requesterUserID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     nil,
			requesterUserID: 3,
			expectedErr:     nil,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			appTestRef, ok := prepareAppTestRef(t, feature.Toggles{
				EnableAuthorization: true,
			})
			if !ok {
				return
			}

			if testCase.prepareData != nil {
				err := testCase.prepareData(appTestRef, testCase.requesterUserID)
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			appName := "Unit Test"
			newApp, internalErr := appTestRef.appService.CreateApp(ct, appName)
			if testCase.expectedErr != nil {
				require.NotNil(t, internalErr)

				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			require.Equal(t, uint64(1), newApp.ID)
			require.Equal(t, testCase.requesterUserID, newApp.CreatorUserID)
			require.Nil(t, newApp.ActiveVersionNumber)
			require.Equal(t, appName, newApp.Name)
			require.Equal(t, uint64(0), newApp.InstallationCount)
			require.Equal(t, "", newApp.Description)
			require.Equal(t, "", newApp.APISecret)
			require.NotNil(t, newApp.CreatedAt)
			require.Nil(t, newApp.UpdatedAt)

			appInDb, internalErr := appTestRef.appDao.FindAppByID(ct, newApp.ID)
			require.Nil(t, internalErr)

			require.True(t, areAppsEqual(newApp, appInDb))

			// (TODO): verify if user group created after we support finding group by resource id in cloud
		})
	}
}

func TestAppService_UpdateApp(t *testing.T) {
	var appID uint64 = 1
	var ownerUserID uint64 = 3
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(appTestRef AppTestRef, requesterUserID uint64, appID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed when user is in app admin group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareAppAdminData,
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "permission denied when user is in app member group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareAppMemberData,
			requesterUserID: 3,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
		{
			name: "permission denied when user is not in any group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     nil,
			requesterUserID: 3,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			appTestRef, ok := prepareAppTestRef(t, feature.Toggles{
				EnableAuthorization: true,
			})
			if !ok {
				return
			}

			if testCase.prepareData != nil {
				err := testCase.prepareData(appTestRef, testCase.requesterUserID, appID)
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			// create app
			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			require.Nil(t, err)

			defer tx.Rollback()
			app := createAppData(appID, nil, ownerUserID)
			require.Nil(t, appTestRef.appDao.CreateApp(ct, tx, app))

			updatedName := "Updated Name"
			updatedDescription := "Updated Description"
			updatedActiveVersionNumber := int32(2)
			input := UpdateAppInput{
				Name:                &updatedName,
				Description:         &updatedDescription,
				ActiveVersionNumber: &updatedActiveVersionNumber,
			}
			updatedApp, internalErr := appTestRef.appService.UpdateApp(ct, appID, input)
			if testCase.expectedErr != nil {
				require.NotNil(t, internalErr)
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			app.ActiveVersionNumber = input.ActiveVersionNumber
			app.Name = *input.Name
			app.Description = *input.Description
			require.True(t, areAppsEqual(app, updatedApp))

			appInDb, internalErr := appTestRef.appDao.FindAppByID(ct, app.ID)
			require.Nil(t, internalErr)
			require.True(t, areAppsEqual(app, appInDb))
		})
	}
}

func TestAppService_DeleteApp(t *testing.T) {
	var appID uint64 = 1
	var ownerUserID uint64 = 3
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(appTestRef AppTestRef, requesterUserID uint64, appID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed when user is in app admin group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareAppAdminData,
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "permission denied when user is in app member group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareAppMemberData,
			requesterUserID: 3,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
		{
			name: "permission denied when user is not in any group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     nil,
			requesterUserID: 3,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			appTestRef, ok := prepareAppTestRef(t, feature.Toggles{
				EnableAuthorization: true,
			})
			if !ok {
				return
			}

			if testCase.prepareData != nil {
				err := testCase.prepareData(appTestRef, testCase.requesterUserID, appID)
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			require.Nil(t, err)

			defer tx.Rollback()
			activeVersionNumber := int32(1)
			app := createAppData(appID, &activeVersionNumber, ownerUserID)
			require.Nil(t, appTestRef.appDao.CreateApp(ct, tx, app))

			deleted, internalErr := appTestRef.appService.DeleteApp(ct, appID)
			if testCase.expectedErr != nil {
				require.NotNil(t, internalErr)

				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			require.True(t, areAppsEqual(deleted, app))

			_, internalErr = appTestRef.appDao.FindAppByID(ct, app.ID)
			require.NotNil(t, internalErr)
			require.Equal(t, internalErr.Code, errs.NotFound)
			// (TODO): verify if user group deleted after we support finding group by resource id in cloud
		})
	}
}

func TestAppService_RefreshAppSecret(t *testing.T) {
	var appID uint64 = 1
	var ownerUserID uint64 = 3
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(appTestRef AppTestRef, requesterUserID uint64, appID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed when user is in app admin group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareAppAdminData,
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "permission denied when user is in app member group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareAppMemberData,
			requesterUserID: 3,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
		{
			name: "permission denied when user is not in any group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     nil,
			requesterUserID: 3,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			appTestRef, ok := prepareAppTestRef(t, feature.Toggles{
				EnableAuthorization: true,
			})
			if !ok {
				return
			}

			if testCase.prepareData != nil {
				err := testCase.prepareData(appTestRef, testCase.requesterUserID, appID)
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			require.Nil(t, err)

			defer tx.Rollback()
			activeVersionNumber := int32(1)
			app := createAppData(appID, &activeVersionNumber, ownerUserID)
			require.Nil(t, appTestRef.appDao.CreateApp(ct, tx, app))

			updated, internalErr := appTestRef.appService.RefreshAppSecret(ct, appID)
			if testCase.expectedErr != nil {
				require.NotNil(t, internalErr)

				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			require.NotEqual(t, app.APISecret, updated.APISecret)
			app.APISecret = updated.APISecret
			require.True(t, areAppsEqual(app, updated))

			appInDb, internalErr := appTestRef.appDao.FindAppByID(ct, app.ID)
			require.Nil(t, internalErr)
			require.True(t, areAppsEqual(app, appInDb))
		})
	}
}

func TestAppService_FindApp(t *testing.T) {
	var appID1 uint64 = 1
	var appID2 uint64 = 2
	var appID3 uint64 = 3
	var ownerUserID uint64 = 3
	var teamID uint64 = 1
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(appTestRef AppTestRef, requesterUserID uint64, appID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     nil,
			requesterUserID: 3,
			expectedErr:     nil,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			appTestRef, ok := prepareAppTestRef(t, feature.Toggles{
				EnableAuthorization: true,
			})
			if !ok {
				return
			}

			if testCase.prepareData != nil {
				err := testCase.prepareData(appTestRef, testCase.requesterUserID, appID1)
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			require.Nil(t, err)

			defer tx.Rollback()
			activeVersionNumber := int32(1)
			app1 := createAppData(appID1, &activeVersionNumber, ownerUserID)
			app2 := createAppData(appID2, &activeVersionNumber, ownerUserID)
			app3 := createAppData(appID3, &activeVersionNumber, ownerUserID)
			appVersion1 := createAppVersionData(appID1, 1, false)
			appVersion2 := createAppVersionData(appID2, 1, false)
			appVersion3 := createAppVersionData(appID3, 1, true)
			appVersionVisibleTeam1 := createAppVersionVisibleTeamData(appID2, 1, teamID)
			require.Nil(t, appTestRef.appDao.CreateApp(ct, tx, app1))
			require.Nil(t, appTestRef.appDao.CreateApp(ct, tx, app2))
			require.Nil(t, appTestRef.appDao.CreateApp(ct, tx, app3))
			require.Nil(t, appTestRef.appVersionDao.CreateAppVersion(ct, tx, appVersion1))
			require.Nil(t, appTestRef.appVersionDao.CreateAppVersion(ct, tx, appVersion2))
			require.Nil(t, appTestRef.appVersionDao.CreateAppVersion(ct, tx, appVersion3))
			require.Nil(t, appTestRef.appVersionVisibleTeamDao.CreateAppVersionVisibleTeam(ct, tx, appVersionVisibleTeam1))

			filter1 := AppFilter{
				AppID: &appID1,
			}
			found, internalErr := appTestRef.appService.FindApps(ct, &filter1)
			require.Nil(t, internalErr)
			require.Equal(t, 1, len(found))
			require.True(t, areAppsEqual(app1, found[0]))

			filter2 := AppFilter{
				TeamID: &teamID,
			}
			found, internalErr = appTestRef.appService.FindApps(ct, &filter2)
			if testCase.expectedErr != nil {
				require.NotNil(t, internalErr)
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			require.Equal(t, 2, len(found))

			for _, app := range found {
				require.True(t, app.ID == appID2 || app.ID == appID3)

				if app.ID == appID2 {
					require.True(t, areAppsEqual(app2, app))
				} else {
					require.True(t, areAppsEqual(app3, app))
				}
			}

			foundApp, internalErr := appTestRef.appService.FindAppByID(ct, appID1)
			if testCase.expectedErr != nil {
				require.NotNil(t, internalErr)
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			require.True(t, areAppsEqual(app1, foundApp))
		})
	}
}

func TestAppService_CreateAppVersion(t *testing.T) {
	var appID uint64 = 1
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(appTestRef AppTestRef, requesterUserID uint64, appID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed when user is in app admin group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareAppAdminData,
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "succeed when user is in app member group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareAppMemberData,
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "permission denied when user is not in any group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     nil,
			requesterUserID: 3,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			appTestRef, ok := prepareAppTestRef(t, feature.Toggles{
				EnableAuthorization: true,
			})
			if !ok {
				return
			}

			if testCase.prepareData != nil {
				err := testCase.prepareData(appTestRef, testCase.requesterUserID, appID)
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			newAppVersion, internalErr := appTestRef.appService.CreateAppVersion(ct, appID)
			if testCase.expectedErr != nil {
				require.NotNil(t, internalErr)
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			require.Equal(t, int32(1), newAppVersion.VersionNumber)
			require.Equal(t, appID, newAppVersion.AppID)
			require.False(t, newAppVersion.IsPublic)
			require.False(t, newAppVersion.HasUIExtension)
			require.Nil(t, newAppVersion.IconURL)
			require.Nil(t, newAppVersion.Changes)
			require.Nil(t, newAppVersion.UIExtensionEntrypointPath)
			require.Nil(t, newAppVersion.UpdateAt)

			appVersionInDb, internalErr := appTestRef.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct,
				newAppVersion.AppID, newAppVersion.VersionNumber)
			require.Nil(t, internalErr)
			require.True(t, areAppVersionsEqual(newAppVersion, appVersionInDb))
		})
	}
}

func TestAppService_UpdateAppVersion(t *testing.T) {
	var appID uint64 = 1
	var versionNumber int32 = 1
	var ownerUserID uint64 = 3
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(appTestRef AppTestRef, requesterUserID uint64, appID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed when user is in app admin group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareAppAdminData,
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "succeed when user is in app member group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareAppMemberData,
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "permission denied when user is not in any group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     nil,
			requesterUserID: 3,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			appTestRef, ok := prepareAppTestRef(t, feature.Toggles{
				EnableAuthorization: true,
			})
			if !ok {
				return
			}

			if testCase.prepareData != nil {
				err := testCase.prepareData(appTestRef, testCase.requesterUserID, appID)
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			require.Nil(t, err)

			defer tx.Rollback()
			app := createAppData(appID, nil, ownerUserID)
			appVersion := createAppVersionData(appID, versionNumber, false)
			require.Nil(t, appTestRef.appDao.CreateApp(ct, tx, app))
			require.Nil(t, appTestRef.appVersionDao.CreateAppVersion(ct, tx, appVersion))

			iconUrl := "Updated URL"
			uiExtensionEntryPointPath := "Updated path"
			changes := "Updated change"
			input := UpdateAppVersionInput{
				IconURL:                   &iconUrl,
				HasUIExtension:            true,
				UIExtensionEntryPointPath: &uiExtensionEntryPointPath,
				Changes:                   &changes,
				IsPublic:                  true,
			}
			updatedAppVersion, internalErr := appTestRef.appService.UpdateAppVersion(ct,
				appVersion.AppID, appVersion.VersionNumber, input)
			if testCase.expectedErr != nil {
				require.NotNil(t, internalErr)
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			appVersion.IconURL = input.IconURL
			appVersion.HasUIExtension = input.HasUIExtension
			appVersion.UIExtensionEntrypointPath = input.UIExtensionEntryPointPath
			appVersion.Changes = input.Changes
			appVersion.IsPublic = input.IsPublic
			require.True(t, areAppVersionsEqual(appVersion, updatedAppVersion))

			appVersionInDb, internalErr := appTestRef.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct,
				appVersion.AppID, appVersion.VersionNumber)
			require.Nil(t, internalErr)
			require.True(t, areAppVersionsEqual(appVersion, appVersionInDb))
		})
	}
}

func TestAppService_DeleteAppVersion(t *testing.T) {
	var appID uint64 = 1
	var versionNumber int32 = 1
	var ownerUserID uint64 = 3
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(appTestRef AppTestRef, requesterUserID uint64, appID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed when user is in app admin group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareAppAdminData,
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "succeed when user is in app member group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareAppMemberData,
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "permission denied when user is not in any group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     nil,
			requesterUserID: 3,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			appTestRef, ok := prepareAppTestRef(t, feature.Toggles{
				EnableAuthorization: true,
			})
			if !ok {
				return
			}

			if testCase.prepareData != nil {
				err := testCase.prepareData(appTestRef, testCase.requesterUserID, appID)
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			require.Nil(t, err)

			defer tx.Rollback()
			app := createAppData(appID, nil, ownerUserID)
			appVersion := createAppVersionData(appID, versionNumber, false)
			require.Nil(t, appTestRef.appDao.CreateApp(ct, tx, app))
			require.Nil(t, appTestRef.appVersionDao.CreateAppVersion(ct, tx, appVersion))

			deletedAppVersion, internalErr := appTestRef.appService.DeleteAppVersion(ct,
				appVersion.AppID, appVersion.VersionNumber)
			if testCase.expectedErr != nil {
				require.NotNil(t, internalErr)
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			require.True(t, areAppVersionsEqual(appVersion, deletedAppVersion))

			_, internalErr = appTestRef.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appVersion.AppID,
				appVersion.VersionNumber)
			require.NotNil(t, internalErr)
			require.Equal(t, internalErr.Code, errs.NotFound)
		})
	}
}

func TestAppService_FindAppVersion(t *testing.T) {
	var appID1 uint64 = 1
	var appID2 uint64 = 2
	var ownerUserID uint64 = 3
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(appTestRef AppTestRef, requesterUserID uint64, appID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     nil,
			requesterUserID: 3,
			expectedErr:     nil,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			appTestRef, ok := prepareAppTestRef(t, feature.Toggles{
				EnableAuthorization: true,
			})
			if !ok {
				return
			}

			if testCase.prepareData != nil {
				err := testCase.prepareData(appTestRef, testCase.requesterUserID, appID1)
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			require.Nil(t, err)

			defer tx.Rollback()
			activeVersionNumber := int32(1)
			app1 := createAppData(appID1, &activeVersionNumber, ownerUserID)
			app2 := createAppData(appID2, &activeVersionNumber, ownerUserID)
			appVersion1 := createAppVersionData(appID1, 1, false)
			appVersion2 := createAppVersionData(appID2, 1, false)
			require.Nil(t, appTestRef.appDao.CreateApp(ct, tx, app1))
			require.Nil(t, appTestRef.appDao.CreateApp(ct, tx, app2))
			require.Nil(t, appTestRef.appVersionDao.CreateAppVersion(ct, tx, appVersion1))
			require.Nil(t, appTestRef.appVersionDao.CreateAppVersion(ct, tx, appVersion2))

			found, internalErr := appTestRef.appService.FindAppVersionByAppID(ct, appID1)
			require.Nil(t, internalErr)
			require.Equal(t, 1, len(found))
			require.True(t, areAppVersionsEqual(appVersion1, found[0]))

			foundVersion, internalErr := appTestRef.appService.FindAppVersionByAppIDAndVersionNumber(ct, appVersion1.AppID,
				appVersion1.VersionNumber)
			require.Nil(t, internalErr)
			require.True(t, areAppVersionsEqual(appVersion1, foundVersion))
		})
	}
}

func TestAppService_CreateAppVersionVisibleTeam(t *testing.T) {
	var appID uint64 = 1
	var versionNumber int32 = 1
	var teamID uint64 = 4
	var ownerUserID uint64 = 3
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(appTestRef AppTestRef, requesterUserID uint64, appID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed when user is in app admin group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareAppAdminData,
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "succeed when user is in app member group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareAppMemberData,
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "permission denied when user is not in any group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     nil,
			requesterUserID: 3,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			appTestRef, ok := prepareAppTestRef(t, feature.Toggles{
				EnableAuthorization: true,
			})
			if !ok {
				return
			}

			if testCase.prepareData != nil {
				err := testCase.prepareData(appTestRef, testCase.requesterUserID, appID)
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			require.Nil(t, err)

			defer tx.Rollback()
			app := createAppData(appID, nil, ownerUserID)
			appVersion := createAppVersionData(appID, versionNumber, false)

			require.Nil(t, appTestRef.appDao.CreateApp(ct, tx, app))
			require.Nil(t, appTestRef.appVersionDao.CreateAppVersion(ct, tx, appVersion))

			returned, internalErr := appTestRef.appService.CreateAppVersionVisibleTeam(ct, appID,
				versionNumber, teamID)
			if testCase.expectedErr != nil {
				require.NotNil(t, internalErr)
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			require.Equal(t, versionNumber, returned.VersionNumber)
			require.Equal(t, appID, returned.AppID)
			require.False(t, returned.IsPublic)
			require.False(t, returned.HasUIExtension)
			require.Nil(t, returned.IconURL)
			require.Nil(t, returned.Changes)
			require.Nil(t, returned.UIExtensionEntrypointPath)

			appVersionVisibleTeamInDb, internalErr := appTestRef.appVersionVisibleTeamDao.
				FindAppVersionVisibleTeamWithTx(ct, tx, appID, versionNumber, teamID)
			require.Nil(t, internalErr)
			require.Equal(t, versionNumber, appVersionVisibleTeamInDb.VersionNumber)
			require.Equal(t, appID, appVersionVisibleTeamInDb.AppID)
			require.Equal(t, teamID, appVersionVisibleTeamInDb.TeamID)
		})
	}
}

func TestAppService_DeleteAppVersionVisibleTeam(t *testing.T) {
	var appID uint64 = 1
	var versionNumber int32 = 1
	var ownerUserID uint64 = 3
	var teamID uint64 = 4
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(appTestRef AppTestRef, requesterUserID uint64, appID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed when user is in app admin group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareAppAdminData,
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "succeed when user is in app member group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareAppMemberData,
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "permission denied when user is not in any group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     nil,
			requesterUserID: 3,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			appTestRef, ok := prepareAppTestRef(t, feature.Toggles{
				EnableAuthorization: true,
			})
			if !ok {
				return
			}

			if testCase.prepareData != nil {
				err := testCase.prepareData(appTestRef, testCase.requesterUserID, appID)
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			require.Nil(t, err)

			defer tx.Rollback()
			app := createAppData(appID, nil, ownerUserID)
			appVersion := createAppVersionData(appID, versionNumber, false)
			appVersionVisibleTeam := createAppVersionVisibleTeamData(appID, versionNumber, teamID)
			require.Nil(t, appTestRef.appDao.CreateApp(ct, tx, app))
			require.Nil(t, appTestRef.appVersionDao.CreateAppVersion(ct, tx, appVersion))
			require.Nil(t, appTestRef.appVersionVisibleTeamDao.CreateAppVersionVisibleTeam(ct, tx,
				appVersionVisibleTeam))

			returned, internalErr := appTestRef.appService.DeleteAppVersionVisibleTeam(ct, appID, versionNumber,
				teamID)
			if testCase.expectedErr != nil {
				require.NotNil(t, internalErr)
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			require.True(t, areAppVersionsEqual(appVersion, returned))

			_, internalErr = appTestRef.appVersionVisibleTeamDao.FindAppVersionVisibleTeamWithTx(ct, tx, appID,
				versionNumber,
				teamID)
			require.NotNil(t, internalErr)
			require.Equal(t, errs.NotFound, internalErr.Code)
		})
	}
}

func TestAppService_FindAppVersionVisibleTeams(t *testing.T) {
	var appID1 uint64 = 1
	var appID2 uint64 = 2
	var versionNumber1 int32 = 1
	var versionNumber2 int32 = 2
	var versionNumber3 int32 = 3
	var teamID uint64 = 4
	var ownerUserID uint64 = 3
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(appTestRef AppTestRef, requesterUserID uint64, appID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     nil,
			requesterUserID: 3,
			expectedErr:     nil,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			appTestRef, ok := prepareAppTestRef(t, feature.Toggles{
				EnableAuthorization: true,
			})
			if !ok {
				return
			}

			if testCase.prepareData != nil {
				err := testCase.prepareData(appTestRef, testCase.requesterUserID, appID1)
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			require.Nil(t, err)

			defer tx.Rollback()
			activeVersionNumber := int32(1)
			app1 := createAppData(appID1, &activeVersionNumber, ownerUserID)
			app2 := createAppData(appID2, &activeVersionNumber, ownerUserID)
			appVersion1 := createAppVersionData(appID1, versionNumber1, false)
			appVersion2 := createAppVersionData(appID2, versionNumber2, false)
			appVersion3 := createAppVersionData(appID2, versionNumber3, false)
			appVersionVisibleTeam := createAppVersionVisibleTeamData(appID2, versionNumber3, teamID)
			team := createTeamData(teamID, ownerUserID)
			require.Nil(t, appTestRef.appDao.CreateApp(ct, tx, app1))
			require.Nil(t, appTestRef.appDao.CreateApp(ct, tx, app2))
			require.Nil(t, appTestRef.appVersionDao.CreateAppVersion(ct, tx, appVersion1))
			require.Nil(t, appTestRef.appVersionDao.CreateAppVersion(ct, tx, appVersion2))
			require.Nil(t, appTestRef.appVersionDao.CreateAppVersion(ct, tx, appVersion3))
			require.Nil(t, appTestRef.appVersionVisibleTeamDao.CreateAppVersionVisibleTeam(ct, tx, appVersionVisibleTeam))
			require.Nil(t, appTestRef.teamDao.CreateTeam(ct, tx, team))

			found, internalErr := appTestRef.appService.FindAppVersionVisibleTeams(ct, appID2, versionNumber3)
			require.Nil(t, internalErr)
			require.Equal(t, 1, len(found))
			require.True(t, areTeamsEqual(team, found[0]))
		})
	}
}

func TestAppService_CreateAppTeamInstallation(t *testing.T) {
	var appID uint64 = 1
	var versionNumber int32 = 1
	var teamID uint64 = 4
	var ownerUserID uint64 = 3
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(appTestRef AppTestRef, requesterUserID uint64, teamID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed when user is in team owner group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareTeamOwnerData,
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "succeed when user is in team admin group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareTeamAdminData,
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "permission denied when user is in team member group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareTeamMemberData,
			requesterUserID: 3,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
		{
			name: "permission denied when user is not in any group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     nil,
			requesterUserID: 3,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			appTestRef, ok := prepareAppTestRef(t, feature.Toggles{
				EnableAuthorization: true,
			})
			if !ok {
				return
			}

			if testCase.prepareData != nil {
				err := testCase.prepareData(appTestRef, testCase.requesterUserID, teamID)
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			require.Nil(t, err)

			defer tx.Rollback()
			app := createAppData(appID, nil, ownerUserID)
			appVersion := createAppVersionData(appID, versionNumber, true)
			team := createTeamData(teamID, ownerUserID)

			require.Nil(t, appTestRef.appDao.CreateApp(ct, tx, app))
			require.Nil(t, appTestRef.appVersionDao.CreateAppVersion(ct, tx, appVersion))
			require.Nil(t, appTestRef.teamDao.CreateTeam(ct, tx, team))

			newAppInstallation, internalErr := appTestRef.appService.CreateAppInstallation(ct, teamID, appID,
				versionNumber)
			if testCase.expectedErr != nil {
				require.NotNil(t, internalErr)
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			require.Equal(t, appID, newAppInstallation.AppID)
			require.Equal(t, versionNumber, newAppInstallation.EnabledVersionNumber)
			require.Equal(t, teamID, newAppInstallation.InstalledTeamID)
			require.Equal(t, ownerUserID, *newAppInstallation.InstalledByUserID)
			require.NotNil(t, newAppInstallation.InstalledAt)

			appInstallationInDb, internalErr := appTestRef.appTeamInstallationDao.
				FindAppTeamInstallationByAppIDAndTeamIDWithTx(ct, tx, appID, teamID)
			require.Nil(t, internalErr)
			require.True(t, areAppInstallationsEqual(appInstallationInDb, newAppInstallation))
		})
	}
}

func TestAppService_UpdateAppTeamInstallation(t *testing.T) {
	var appID uint64 = 1
	var versionNumber1 int32 = 1
	var versionNumber2 int32 = 2
	var teamID uint64 = 4
	var ownerUserID uint64 = 3
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(appTestRef AppTestRef, requesterUserID uint64, teamID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed when user is in team owner group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareTeamOwnerData,
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "succeed when user is in team admin group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareTeamAdminData,
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "permission denied when user is in team member group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareTeamMemberData,
			requesterUserID: 3,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
		{
			name: "permission denied when user is not in any group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     nil,
			requesterUserID: 3,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			appTestRef, ok := prepareAppTestRef(t, feature.Toggles{
				EnableAuthorization: true,
			})
			if !ok {
				return
			}

			if testCase.prepareData != nil {
				err := testCase.prepareData(appTestRef, testCase.requesterUserID, teamID)
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			require.Nil(t, err)

			defer tx.Rollback()
			app := createAppData(appID, nil, ownerUserID)
			appVersion1 := createAppVersionData(appID, versionNumber1, true)
			appVersion2 := createAppVersionData(appID, versionNumber2, true)
			team := createTeamData(teamID, ownerUserID)
			appTeamInstallation := createAppTeamInstallation(appID, versionNumber1, teamID, ownerUserID)

			require.Nil(t, appTestRef.appDao.CreateApp(ct, tx, app))
			require.Nil(t, appTestRef.appVersionDao.CreateAppVersion(ct, tx, appVersion1))
			require.Nil(t, appTestRef.appVersionDao.CreateAppVersion(ct, tx, appVersion2))
			require.Nil(t, appTestRef.teamDao.CreateTeam(ct, tx, team))
			require.Nil(t, appTestRef.appTeamInstallationDao.CreateAppTeamInstallation(ct, tx, appTeamInstallation))

			input := UpdateAppTeamInstallationInput{
				EnabledVersionNumber: versionNumber2,
			}
			updated, internalErr := appTestRef.appService.UpdateAppInstallation(ct, appID, teamID,
				input)
			if testCase.expectedErr != nil {
				require.NotNil(t, internalErr)
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			appTeamInstallation.EnabledVersionNumber = versionNumber2
			require.True(t, areAppInstallationsEqual(appTeamInstallation, updated))

			appInstallationInDb, internalErr := appTestRef.appTeamInstallationDao.
				FindAppTeamInstallationByAppIDAndTeamIDWithTx(ct, tx, appID, teamID)
			require.Nil(t, internalErr)
			require.True(t, areAppInstallationsEqual(appInstallationInDb, appTeamInstallation))
		})
	}
}

func TestAppService_DeleteAppTeamInstallation(t *testing.T) {
	var appID uint64 = 1
	var versionNumber int32 = 1
	var teamID uint64 = 4
	var ownerUserID uint64 = 3
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(appTestRef AppTestRef, requesterUserID uint64, teamID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed when user is in team owner group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareTeamOwnerData,
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "succeed when user is in team admin group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareTeamAdminData,
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "permission denied when user is in team member group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     prepareTeamMemberData,
			requesterUserID: 3,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
		{
			name: "permission denied when user is not in any group",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     nil,
			requesterUserID: 3,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			appTestRef, ok := prepareAppTestRef(t, feature.Toggles{
				EnableAuthorization: true,
			})
			if !ok {
				return
			}

			if testCase.prepareData != nil {
				err := testCase.prepareData(appTestRef, testCase.requesterUserID, teamID)
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			require.Nil(t, err)

			defer tx.Rollback()
			app := createAppData(appID, nil, ownerUserID)
			appVersion := createAppVersionData(appID, versionNumber, true)
			team := createTeamData(teamID, ownerUserID)
			appTeamInstallation := createAppTeamInstallation(appID, versionNumber, teamID, ownerUserID)

			require.Nil(t, appTestRef.appDao.CreateApp(ct, tx, app))
			require.Nil(t, appTestRef.appVersionDao.CreateAppVersion(ct, tx, appVersion))
			require.Nil(t, appTestRef.teamDao.CreateTeam(ct, tx, team))
			require.Nil(t, appTestRef.appTeamInstallationDao.CreateAppTeamInstallation(ct, tx, appTeamInstallation))

			deleted, internalErr := appTestRef.appService.DeleteAppInstallation(ct, appID, teamID)
			if testCase.expectedErr != nil {
				require.NotNil(t, internalErr)
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			require.True(t, areAppInstallationsEqual(appTeamInstallation, deleted))

			_, internalErr = appTestRef.appTeamInstallationDao.
				FindAppTeamInstallationByAppIDAndTeamIDWithTx(ct, tx, appID, teamID)
			require.NotNil(t, internalErr)
			require.Equal(t, errs.NotFound, internalErr.Code)
		})
	}
}

func TestAppService_FindAppTeamInstallations(t *testing.T) {
	var appID1 uint64 = 1
	var appID2 uint64 = 2
	var versionNumber int32 = 1
	var teamID1 uint64 = 4
	var teamID2 uint64 = 5
	var ownerUserID uint64 = 3
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(appTestRef AppTestRef, requesterUserID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData:     nil,
			requesterUserID: 3,
			expectedErr:     nil,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			appTestRef, ok := prepareAppTestRef(t, feature.Toggles{
				EnableAuthorization: true,
			})
			if !ok {
				return
			}

			if testCase.prepareData != nil {
				err := testCase.prepareData(appTestRef, testCase.requesterUserID)
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			require.Nil(t, err)

			defer tx.Rollback()
			app1 := createAppData(appID1, nil, ownerUserID)
			app2 := createAppData(appID2, nil, ownerUserID)
			appVersion1 := createAppVersionData(appID1, versionNumber, true)
			appVersion2 := createAppVersionData(appID2, versionNumber, true)
			team1 := createTeamData(teamID1, ownerUserID)
			team2 := createTeamData(teamID2, ownerUserID)
			appTeamInstallation1 := createAppTeamInstallation(appID1, versionNumber, teamID1, ownerUserID)
			appTeamInstallation2 := createAppTeamInstallation(appID2, versionNumber, teamID2, ownerUserID)
			appTeamInstallation3 := createAppTeamInstallation(appID1, versionNumber, teamID2, ownerUserID)

			require.Nil(t, appTestRef.appDao.CreateApp(ct, tx, app1))
			require.Nil(t, appTestRef.appDao.CreateApp(ct, tx, app2))
			require.Nil(t, appTestRef.appVersionDao.CreateAppVersion(ct, tx, appVersion1))
			require.Nil(t, appTestRef.appVersionDao.CreateAppVersion(ct, tx, appVersion2))
			require.Nil(t, appTestRef.teamDao.CreateTeam(ct, tx, team1))
			require.Nil(t, appTestRef.teamDao.CreateTeam(ct, tx, team2))
			require.Nil(t, appTestRef.appTeamInstallationDao.CreateAppTeamInstallation(ct, tx, appTeamInstallation1))
			require.Nil(t, appTestRef.appTeamInstallationDao.CreateAppTeamInstallation(ct, tx, appTeamInstallation2))
			require.Nil(t, appTestRef.appTeamInstallationDao.CreateAppTeamInstallation(ct, tx, appTeamInstallation3))

			found, internalErr := appTestRef.appService.FindAppTeamInstallationsByAppID(ct, appID1)
			require.Nil(t, internalErr)

			require.Equal(t, 2, len(found))
			for _, appTeamInstallation := range found {
				require.True(t, appTeamInstallation.InstalledTeamID == teamID1 ||
					appTeamInstallation.InstalledTeamID == teamID2)

				if appTeamInstallation.InstalledTeamID == teamID1 {
					require.True(t, areAppInstallationsEqual(appTeamInstallation, appTeamInstallation1))
				} else {
					require.True(t, areAppInstallationsEqual(appTeamInstallation, appTeamInstallation3))
				}
			}

			found, internalErr = appTestRef.appService.FindAppInstallationsByTeamID(ct, teamID2)
			require.Nil(t, internalErr)
			require.Equal(t, 2, len(found))

			for _, appTeamInstallation := range found {
				require.True(t, appTeamInstallation.AppID == appID1 ||
					appTeamInstallation.AppID == appID2)

				if appTeamInstallation.AppID == appID1 {
					require.True(t, areAppInstallationsEqual(appTeamInstallation, appTeamInstallation3))
				} else {
					require.True(t, areAppInstallationsEqual(appTeamInstallation, appTeamInstallation2))
				}
			}
		})
	}
}

func prepareAppAdminData(appTestRef AppTestRef, requesterUserID uint64, appID uint64) *errs.Error {
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, 1)
	group, err := appTestRef.
		cloudTestKit.
		AuthorizationService.
		CreateUserGroup(ct, "Admin", nil)
	if err != nil {
		return err
	}

	return servicetest.AddAppPermission(
		ct,
		appTestRef.cloudTestKit.AuthorizationService,
		appID,
		group.ID,
		authorization.AppAdminResourceTypeOperations,
		requesterUserID)
}

func prepareAppMemberData(appTestRef AppTestRef, requesterUserID uint64, appID uint64) *errs.Error {
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, 1)
	group, err := appTestRef.
		cloudTestKit.
		AuthorizationService.
		CreateUserGroup(ct, "Member", nil)
	if err != nil {
		return err
	}

	return servicetest.AddAppPermission(
		ct,
		appTestRef.cloudTestKit.AuthorizationService,
		appID,
		group.ID,
		authorization.AppMemberResourceTypeOperations,
		requesterUserID)
}

func prepareTeamOwnerData(appTestRef AppTestRef, requesterUserID uint64, teamID uint64) *errs.Error {
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, 1)
	group, err := appTestRef.
		cloudTestKit.
		AuthorizationService.
		CreateUserGroup(ct, "TeamOwner", nil)
	if err != nil {
		return err
	}

	return servicetest.AddTeamPermission(
		ct,
		appTestRef.cloudTestKit.AuthorizationService,
		teamID,
		group.ID,
		authorization.TeamOwnerResourceTypeOperations,
		requesterUserID)
}

func prepareTeamAdminData(appTestRef AppTestRef, requesterUserID uint64, teamID uint64) *errs.Error {
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, 1)
	group, err := appTestRef.
		cloudTestKit.
		AuthorizationService.
		CreateUserGroup(ct, "TeamAdmin", nil)
	if err != nil {
		return err
	}

	return servicetest.AddTeamPermission(
		ct,
		appTestRef.cloudTestKit.AuthorizationService,
		teamID,
		group.ID,
		authorization.TeamAdminResourceTypeOperations,
		requesterUserID)
}

func prepareTeamMemberData(appTestRef AppTestRef, requesterUserID uint64, teamID uint64) *errs.Error {
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, 1)
	group, err := appTestRef.
		cloudTestKit.
		AuthorizationService.
		CreateUserGroup(ct, "TeamMember", nil)
	if err != nil {
		return err
	}

	return servicetest.AddTeamPermission(
		ct,
		appTestRef.cloudTestKit.AuthorizationService,
		teamID,
		group.ID,
		authorization.TeamMemberResourceTypeOperations,
		requesterUserID)
}

func areAppsEqual(one entity.App, other entity.App) bool {
	if one.ID != other.ID {
		return false
	}

	if one.Name != other.Name {
		return false
	}

	if one.APISecret != other.APISecret {
		return false
	}

	if one.Description != other.Description {
		return false
	}

	if one.CreatedAt != other.CreatedAt {
		return false
	}

	if one.ActiveVersionNumber == nil || other.ActiveVersionNumber == nil {
		if one.ActiveVersionNumber != nil || other.ActiveVersionNumber != nil {
			return false
		}
	} else if *one.ActiveVersionNumber != *other.ActiveVersionNumber {
		return false
	}

	if one.InstallationCount != other.InstallationCount {
		return false
	}

	if one.CreatorUserID != other.CreatorUserID {
		return false
	}

	return true
}

func areAppVersionsEqual(one entity.AppVersion, other entity.AppVersion) bool {
	if one.AppID != other.AppID {
		return false
	}

	if one.VersionNumber != other.VersionNumber {
		return false
	}

	if one.IsPublic != other.IsPublic {
		return false
	}

	if one.HasUIExtension != other.HasUIExtension {
		return false
	}

	if one.CreatedAt != other.CreatedAt {
		return false
	}

	if one.UIExtensionEntrypointPath == nil || other.UIExtensionEntrypointPath == nil {
		if one.UIExtensionEntrypointPath != nil || other.UIExtensionEntrypointPath != nil {
			return false
		}
	} else if *one.UIExtensionEntrypointPath != *other.UIExtensionEntrypointPath {
		return false
	}

	if one.IconURL == nil || other.IconURL == nil {
		if one.IconURL != nil || other.IconURL != nil {
			return false
		}
	} else if *one.IconURL != *other.IconURL {
		return false
	}

	if one.Changes == nil || other.Changes == nil {
		if one.Changes != nil || other.Changes != nil {
			return false
		}
	} else if *one.Changes != *other.Changes {
		return false
	}

	return true
}

func areTeamsEqual(one entity.Team, other entity.Team) bool {
	if one.ID != other.ID {
		return false
	}

	if one.Name != other.Name {
		return false
	}

	if one.CreatorUserID != other.CreatorUserID {
		return false
	}

	if one.OwnerUserID != other.OwnerUserID {
		return false
	}

	if one.CreatedAt != other.CreatedAt {
		return false
	}

	if one.IconURL == nil || other.IconURL == nil {
		if one.IconURL != nil || other.IconURL != nil {
			return false
		}
	} else if *one.IconURL != *other.IconURL {
		return false
	}

	return true
}

func areAppInstallationsEqual(one entity.AppTeamInstallation, other entity.AppTeamInstallation) bool {
	if one.AppID != other.AppID {
		return false
	}

	if one.EnabledVersionNumber != other.EnabledVersionNumber {
		return false
	}

	if one.InstalledTeamID != other.InstalledTeamID {
		return false
	}

	if one.InstalledAt != other.InstalledAt {
		return false
	}

	if one.InstalledByUserID == nil || other.InstalledByUserID == nil {
		if one.InstalledByUserID != nil || other.InstalledByUserID != nil {
			return false
		}
	} else if *one.InstalledByUserID != *other.InstalledByUserID {
		return false
	}

	return true
}

func createAppData(appID uint64, activeVersionNumber *int32, ownerUserID uint64) entity.App {
	return entity.App{
		ID:                  appID,
		Name:                "Unit Test",
		Description:         "Test Description",
		APISecret:           "Test Secret",
		ActiveVersionNumber: activeVersionNumber,
		InstallationCount:   0,
		CreatorUserID:       ownerUserID,
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           nil,
	}
}

func createAppVersionData(appID uint64, versionNumber int32, isPublic bool) entity.AppVersion {
	return entity.AppVersion{
		AppID:                     appID,
		VersionNumber:             versionNumber,
		IconURL:                   nil,
		HasUIExtension:            false,
		UIExtensionEntrypointPath: nil,
		IsPublic:                  isPublic,
		Changes:                   nil,
		CreatedAt:                 time.Now().UTC(),
		UpdateAt:                  nil,
	}
}

func createAppVersionVisibleTeamData(appID uint64, versionNumber int32, teamID uint64) entity.AppVersionVisibleTeam {
	return entity.AppVersionVisibleTeam{
		AppID:         appID,
		VersionNumber: versionNumber,
		TeamID:        teamID,
	}
}

func createTeamData(teamID uint64, userID uint64) entity.Team {
	return entity.Team{
		ID:            teamID,
		Name:          "Unit Test",
		IconURL:       nil,
		CreatorUserID: userID,
		OwnerUserID:   userID,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     nil,
	}
}

func createAppTeamInstallation(appID uint64, versionNumber int32, teamID uint64, userID uint64) entity.AppTeamInstallation {
	return entity.AppTeamInstallation{
		AppID:                appID,
		InstalledTeamID:      teamID,
		InstalledByUserID:    &userID,
		EnabledVersionNumber: versionNumber,
		InstalledAt:          time.Now().UTC(),
	}
}
