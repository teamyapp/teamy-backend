package service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/cloud/testkit"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/daov2/daotestv2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/service/servicetest"
)

type AppTestRef struct {
	appService                 App
	appDaoV2                   daov2.App
	appVersionDaoV2            daov2.AppVersion
	appVersionVisibleTeamDaoV2 daov2.AppVersionVisibleTeam
	teamDaoV2                  daov2.Team
	appTeamInstallationDaoV2   daov2.AppTeamInstallation
	transactionFactory         transaction.Factory
	cloudTestKit               testkit.TestKit
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
	if !assert.Nil(t, internalErr) {
		return AppTestRef{}, false
	}

	ct := context.Background()
	var accountOwner uint64 = 0
	serviceAccountID, internalErr := cloudTestKit.IdentityService.CreateServiceAccount(ct, accountOwner, "test")
	if !assert.Nil(t, internalErr) {
		return AppTestRef{}, false
	}

	apiToken, internalErr := cloudTestKit.IdentityService.GenerateServiceToken(ct, accountOwner, serviceAccountID)
	if !assert.Nil(t, internalErr) {
		return AppTestRef{}, false
	}

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
	if !assert.Nil(t, err) {
		return AppTestRef{}, false
	}

	authorizer := cloudClient.NewAuthorizer(logger, cloudClientRegistry)

	transactionFactory := transaction.NewFactory(nil)

	teamyBackendDB := dbtest.NewInMemoryDB()
	teamyBackendDB.CreateTable(daotestv2.AppTableName)
	teamyBackendDB.CreateTable(daotestv2.AppVersionTableName)
	teamyBackendDB.CreateTable(daotestv2.AppVersionVisibleTeamTableName)
	teamyBackendDB.CreateTable(daotestv2.AppTeamInstallationTableName)
	teamyBackendDB.CreateTable(daotestv2.TeamTableName)

	teamMemberDaoV2 := daotestv2.NewTeamMember(teamyBackendDB, transactionFactory)
	stateSyncer := realtime.NewStateSyncer(logger, teamMemberDaoV2)

	appDaoV2 := daotestv2.NewApp(teamyBackendDB, transactionFactory)
	appVersionDaoV2 := daotestv2.NewAppVersion(teamyBackendDB, transactionFactory)
	appVersionVisibleTeamDaoV2 := daotestv2.NewAppVersionVisibleTeam(teamyBackendDB, transactionFactory)
	appTeamInstallationDaoV2 := daotestv2.NewAppTeamInstallation(teamyBackendDB, transactionFactory)
	teamDaoV2 := daotestv2.NewTeam(teamyBackendDB, transactionFactory)

	appService := NewApp(
		logger,
		cloudClientRegistry,
		authorizer,
		toggles,
		transactionFactory,
		stateSyncer,
		appDaoV2,
		appVersionDaoV2,
		appTeamInstallationDaoV2,
		appVersionVisibleTeamDaoV2,
		teamDaoV2,
	)

	return AppTestRef{
		appService:                 appService,
		transactionFactory:         transactionFactory,
		appDaoV2:                   appDaoV2,
		appVersionDaoV2:            appVersionDaoV2,
		appVersionVisibleTeamDaoV2: appVersionVisibleTeamDaoV2,
		teamDaoV2:                  teamDaoV2,
		appTeamInstallationDaoV2:   appTeamInstallationDaoV2,
		cloudTestKit:               cloudTestKit,
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
				if !assert.Nil(t, err) {
					return
				}
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			appName := "Unit Test"
			newApp, internalErr := appTestRef.appService.CreateApp(ct, appName)
			if testCase.expectedErr != nil {
				if !assert.NotNil(t, internalErr) {
					return
				}

				assert.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else if !assert.Nil(t, internalErr) {
				return
			}

			assert.Equal(t, uint64(1), newApp.ID)
			assert.Equal(t, testCase.requesterUserID, newApp.CreatorUserID)
			assert.Nil(t, newApp.ActiveVersionNumber)
			assert.Equal(t, appName, newApp.Name)
			assert.Equal(t, uint64(0), newApp.InstallationCount)
			assert.Equal(t, "", newApp.Description)
			assert.Equal(t, "", newApp.APISecret)
			assert.NotNil(t, newApp.CreatedAt)
			assert.Nil(t, newApp.UpdatedAt)

			// verify data in db
			appInDb, internalErr := appTestRef.appDaoV2.FindAppByID(ct, newApp.ID)
			if !assert.Nil(t, internalErr) {
				return
			}

			assert.True(t, areAppsEqual(newApp, appInDb))

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
				if !assert.Nil(t, err) {
					return
				}
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			// create app
			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			if !assert.Nil(t, err) {
				return
			}

			defer tx.Rollback()
			app := createAppData(appID, nil, ownerUserID)
			if !assert.Nil(t, appTestRef.appDaoV2.CreateApp(ct, tx, app)) {
				return
			}

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
				if !assert.NotNil(t, internalErr) {
					return
				}

				assert.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else if !assert.Nil(t, internalErr) {
				return
			}

			app.ActiveVersionNumber = input.ActiveVersionNumber
			app.Name = *input.Name
			app.Description = *input.Description
			assert.True(t, areAppsEqual(app, updatedApp))

			// verify data in db
			appInDb, internalErr := appTestRef.appDaoV2.FindAppByID(ct, app.ID)
			if !assert.Nil(t, internalErr) {
				return
			}

			assert.True(t, areAppsEqual(app, appInDb))
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
				if !assert.Nil(t, err) {
					return
				}
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			// create app
			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			if !assert.Nil(t, err) {
				return
			}

			defer tx.Rollback()
			activeVersionNumber := int32(1)
			app := createAppData(appID, &activeVersionNumber, ownerUserID)
			if !assert.Nil(t, appTestRef.appDaoV2.CreateApp(ct, tx, app)) {
				return
			}

			deleted, internalErr := appTestRef.appService.DeleteApp(ct, appID)
			if testCase.expectedErr != nil {
				if !assert.NotNil(t, internalErr) {
					return
				}

				assert.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else if !assert.Nil(t, internalErr) {
				return
			}

			assert.True(t, areAppsEqual(deleted, app))

			// verify data in db
			_, internalErr = appTestRef.appDaoV2.FindAppByID(ct, app.ID)
			if !assert.NotNil(t, internalErr) {
				return
			}

			assert.Equal(t, internalErr.Code, errs.NotFound)

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
				if !assert.Nil(t, err) {
					return
				}
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			// create app
			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			if !assert.Nil(t, err) {
				return
			}

			defer tx.Rollback()
			activeVersionNumber := int32(1)
			app := createAppData(appID, &activeVersionNumber, ownerUserID)
			if !assert.Nil(t, appTestRef.appDaoV2.CreateApp(ct, tx, app)) {
				return
			}

			updated, internalErr := appTestRef.appService.RefreshAppSecret(ct, appID)
			if testCase.expectedErr != nil {
				if !assert.NotNil(t, internalErr) {
					return
				}

				assert.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else if !assert.Nil(t, internalErr) {
				return
			}

			assert.NotEqual(t, app.APISecret, updated.APISecret)
			app.APISecret = updated.APISecret
			assert.True(t, areAppsEqual(app, updated))

			// verify data in db
			appInDb, internalErr := appTestRef.appDaoV2.FindAppByID(ct, app.ID)
			if !assert.Nil(t, internalErr) {
				return
			}

			assert.True(t, areAppsEqual(app, appInDb))
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
				if !assert.Nil(t, err) {
					return
				}
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			// create app
			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			if !assert.Nil(t, err) {
				return
			}

			defer tx.Rollback()
			activeVersionNumber := int32(1)
			app1 := createAppData(appID1, &activeVersionNumber, ownerUserID)
			app2 := createAppData(appID2, &activeVersionNumber, ownerUserID)
			app3 := createAppData(appID3, &activeVersionNumber, ownerUserID)
			appVersion1 := createAppVersionData(appID1, 1, false)
			appVersion2 := createAppVersionData(appID2, 1, false)
			appVersion3 := createAppVersionData(appID3, 1, true)
			appVersionVisibleTeam1 := createAppVersionVisibleTeamData(appID2, 1, teamID)
			if !assert.Nil(t, appTestRef.appDaoV2.CreateApp(ct, tx, app1)) {
				return
			}

			if !assert.Nil(t, appTestRef.appDaoV2.CreateApp(ct, tx, app2)) {
				return
			}

			if !assert.Nil(t, appTestRef.appDaoV2.CreateApp(ct, tx, app3)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionDaoV2.CreateAppVersion(ct, tx, appVersion1)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionDaoV2.CreateAppVersion(ct, tx, appVersion2)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionDaoV2.CreateAppVersion(ct, tx, appVersion3)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionVisibleTeamDaoV2.CreateAppVersionVisibleTeam(ct, tx, appVersionVisibleTeam1)) {
				return
			}

			filter1 := AppFilter{
				AppID: &appID1,
			}
			found, internalErr := appTestRef.appService.FindApps(ct, &filter1)
			if !assert.Nil(t, internalErr) {
				return
			}

			if !assert.Equal(t, 1, len(found)) {
				return
			}

			assert.True(t, areAppsEqual(app1, found[0]))

			filter2 := AppFilter{
				TeamID: &teamID,
			}
			found, internalErr = appTestRef.appService.FindApps(ct, &filter2)
			if testCase.expectedErr != nil {
				if !assert.NotNil(t, internalErr) {
					return
				}

				assert.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else if !assert.Nil(t, internalErr) {
				return
			}

			if !assert.Equal(t, 2, len(found)) {
				return
			}

			for _, app := range found {
				if !assert.True(t, app.ID == appID2 || app.ID == appID3) {
					return
				}

				if app.ID == appID2 {
					assert.True(t, areAppsEqual(app2, app))
				} else {
					assert.True(t, areAppsEqual(app3, app))
				}
			}

			foundApp, internalErr := appTestRef.appService.FindAppByID(ct, appID1)
			if testCase.expectedErr != nil {
				if !assert.NotNil(t, internalErr) {
					return
				}

				assert.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else if !assert.Nil(t, internalErr) {
				return
			}

			assert.True(t, areAppsEqual(app1, foundApp))
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
				if !assert.Nil(t, err) {
					return
				}
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			newAppVersion, internalErr := appTestRef.appService.CreateAppVersion(ct, appID)
			if testCase.expectedErr != nil {
				if !assert.NotNil(t, internalErr) {
					return
				}

				assert.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else if !assert.Nil(t, internalErr) {
				return
			}

			assert.Equal(t, int32(1), newAppVersion.VersionNumber)
			assert.Equal(t, appID, newAppVersion.AppID)
			assert.False(t, newAppVersion.IsPublic)
			assert.False(t, newAppVersion.HasUIExtension)
			assert.Nil(t, newAppVersion.IconURL)
			assert.Nil(t, newAppVersion.Changes)
			assert.Nil(t, newAppVersion.UIExtensionEntrypointPath)
			assert.Nil(t, newAppVersion.UpdateAt)

			// verify data in db
			appVersionInDb, internalErr := appTestRef.appVersionDaoV2.FindAppVersionByAppIDAndVersionNumber(ct,
				newAppVersion.AppID, newAppVersion.VersionNumber)
			if !assert.Nil(t, internalErr) {
				return
			}

			assert.True(t, areAppVersionsEqual(newAppVersion, appVersionInDb))
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
				if !assert.Nil(t, err) {
					return
				}
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			// create app and app version
			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			if !assert.Nil(t, err) {
				return
			}

			defer tx.Rollback()
			app := createAppData(appID, nil, ownerUserID)
			appVersion := createAppVersionData(appID, versionNumber, false)
			if !assert.Nil(t, appTestRef.appDaoV2.CreateApp(ct, tx, app)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionDaoV2.CreateAppVersion(ct, tx, appVersion)) {
				return
			}

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
				if !assert.NotNil(t, internalErr) {
					return
				}

				assert.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else if !assert.Nil(t, internalErr) {
				return
			}

			appVersion.IconURL = input.IconURL
			appVersion.HasUIExtension = input.HasUIExtension
			appVersion.UIExtensionEntrypointPath = input.UIExtensionEntryPointPath
			appVersion.Changes = input.Changes
			appVersion.IsPublic = input.IsPublic
			assert.True(t, areAppVersionsEqual(appVersion, updatedAppVersion))

			// verify data in db
			appVersionInDb, internalErr := appTestRef.appVersionDaoV2.FindAppVersionByAppIDAndVersionNumber(ct,
				appVersion.AppID, appVersion.VersionNumber)
			if !assert.Nil(t, internalErr) {
				return
			}

			assert.True(t, areAppVersionsEqual(appVersion, appVersionInDb))
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
				if !assert.Nil(t, err) {
					return
				}
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			// create app and app version
			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			if !assert.Nil(t, err) {
				return
			}

			defer tx.Rollback()
			app := createAppData(appID, nil, ownerUserID)
			appVersion := createAppVersionData(appID, versionNumber, false)
			if !assert.Nil(t, appTestRef.appDaoV2.CreateApp(ct, tx, app)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionDaoV2.CreateAppVersion(ct, tx, appVersion)) {
				return
			}

			deletedAppVersion, internalErr := appTestRef.appService.DeleteAppVersion(ct,
				appVersion.AppID, appVersion.VersionNumber)
			if testCase.expectedErr != nil {
				if !assert.NotNil(t, internalErr) {
					return
				}

				assert.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else if !assert.Nil(t, internalErr) {
				return
			}

			assert.True(t, areAppVersionsEqual(appVersion, deletedAppVersion))

			// verify data in db
			_, internalErr = appTestRef.appVersionDaoV2.FindAppVersionByAppIDAndVersionNumber(ct, appVersion.AppID,
				appVersion.VersionNumber)
			if !assert.NotNil(t, internalErr) {
				return
			}

			assert.Equal(t, internalErr.Code, errs.NotFound)
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
				if !assert.Nil(t, err) {
					return
				}
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			// create app
			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			if !assert.Nil(t, err) {
				return
			}

			defer tx.Rollback()
			activeVersionNumber := int32(1)
			app1 := createAppData(appID1, &activeVersionNumber, ownerUserID)
			app2 := createAppData(appID2, &activeVersionNumber, ownerUserID)
			appVersion1 := createAppVersionData(appID1, 1, false)
			appVersion2 := createAppVersionData(appID2, 1, false)
			if !assert.Nil(t, appTestRef.appDaoV2.CreateApp(ct, tx, app1)) {
				return
			}

			if !assert.Nil(t, appTestRef.appDaoV2.CreateApp(ct, tx, app2)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionDaoV2.CreateAppVersion(ct, tx, appVersion1)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionDaoV2.CreateAppVersion(ct, tx, appVersion2)) {
				return
			}

			found, internalErr := appTestRef.appService.FindAppVersionByAppID(ct, appID1)
			if !assert.Nil(t, internalErr) {
				return
			}

			if !assert.Equal(t, 1, len(found)) {
				return
			}

			assert.True(t, areAppVersionsEqual(appVersion1, found[0]))

			foundVersion, internalErr := appTestRef.appService.FindAppVersionByAppIDAndVersionNumber(ct, appVersion1.AppID,
				appVersion1.VersionNumber)
			if !assert.Nil(t, internalErr) {
				return
			}

			assert.True(t, areAppVersionsEqual(appVersion1, foundVersion))
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
				if !assert.Nil(t, err) {
					return
				}
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			// create app
			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			if !assert.Nil(t, err) {
				return
			}

			defer tx.Rollback()
			app := createAppData(appID, nil, ownerUserID)
			appVersion := createAppVersionData(appID, versionNumber, false)

			if !assert.Nil(t, appTestRef.appDaoV2.CreateApp(ct, tx, app)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionDaoV2.CreateAppVersion(ct, tx, appVersion)) {
				return
			}

			returned, internalErr := appTestRef.appService.CreateAppVersionVisibleTeam(ct, appID,
				versionNumber, teamID)
			if testCase.expectedErr != nil {
				if !assert.NotNil(t, internalErr) {
					return
				}

				assert.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else if !assert.Nil(t, internalErr) {
				return
			}

			assert.Equal(t, versionNumber, returned.VersionNumber)
			assert.Equal(t, appID, returned.AppID)
			assert.False(t, returned.IsPublic)
			assert.False(t, returned.HasUIExtension)
			assert.Nil(t, returned.IconURL)
			assert.Nil(t, returned.Changes)
			assert.Nil(t, returned.UIExtensionEntrypointPath)

			// verify data in db
			appVersionVisibleTeamInDb, internalErr := appTestRef.appVersionVisibleTeamDaoV2.
				FindAppVersionVisibleTeamWithTx(ct, tx, appID, versionNumber, teamID)
			if !assert.Nil(t, internalErr) {
				return
			}

			assert.Equal(t, versionNumber, appVersionVisibleTeamInDb.VersionNumber)
			assert.Equal(t, appID, appVersionVisibleTeamInDb.AppID)
			assert.Equal(t, teamID, appVersionVisibleTeamInDb.TeamID)
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
				if !assert.Nil(t, err) {
					return
				}
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			// create app and app version
			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			if !assert.Nil(t, err) {
				return
			}

			defer tx.Rollback()
			app := createAppData(appID, nil, ownerUserID)
			appVersion := createAppVersionData(appID, versionNumber, false)
			appVersionVisibleTeam := createAppVersionVisibleTeamData(appID, versionNumber, teamID)
			if !assert.Nil(t, appTestRef.appDaoV2.CreateApp(ct, tx, app)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionDaoV2.CreateAppVersion(ct, tx, appVersion)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionVisibleTeamDaoV2.CreateAppVersionVisibleTeam(ct, tx,
				appVersionVisibleTeam)) {
				return
			}

			returned, internalErr := appTestRef.appService.DeleteAppVersionVisibleTeam(ct, appID, versionNumber,
				teamID)
			if testCase.expectedErr != nil {
				if !assert.NotNil(t, internalErr) {
					return
				}

				assert.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else if !assert.Nil(t, internalErr) {
				return
			}

			assert.True(t, areAppVersionsEqual(appVersion, returned))

			// verify data in db
			_, internalErr = appTestRef.appVersionVisibleTeamDaoV2.FindAppVersionVisibleTeamWithTx(ct, tx, appID,
				versionNumber,
				teamID)
			if !assert.NotNil(t, internalErr) {
				return
			}

			assert.Equal(t, errs.NotFound, internalErr.Code)
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
				if !assert.Nil(t, err) {
					return
				}
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			// create app
			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			if !assert.Nil(t, err) {
				return
			}

			defer tx.Rollback()
			activeVersionNumber := int32(1)
			app1 := createAppData(appID1, &activeVersionNumber, ownerUserID)
			app2 := createAppData(appID2, &activeVersionNumber, ownerUserID)
			appVersion1 := createAppVersionData(appID1, versionNumber1, false)
			appVersion2 := createAppVersionData(appID2, versionNumber2, false)
			appVersion3 := createAppVersionData(appID2, versionNumber3, false)
			appVersionVisibleTeam := createAppVersionVisibleTeamData(appID2, versionNumber3, teamID)
			team := createTeamData(teamID, ownerUserID)
			if !assert.Nil(t, appTestRef.appDaoV2.CreateApp(ct, tx, app1)) {
				return
			}

			if !assert.Nil(t, appTestRef.appDaoV2.CreateApp(ct, tx, app2)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionDaoV2.CreateAppVersion(ct, tx, appVersion1)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionDaoV2.CreateAppVersion(ct, tx, appVersion2)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionDaoV2.CreateAppVersion(ct, tx, appVersion3)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionVisibleTeamDaoV2.CreateAppVersionVisibleTeam(ct, tx, appVersionVisibleTeam)) {
				return
			}

			if !assert.Nil(t, appTestRef.teamDaoV2.CreateTeam(ct, tx, team)) {
				return
			}

			found, internalErr := appTestRef.appService.FindAppVersionVisibleTeams(ct, appID2, versionNumber3)
			if !assert.Nil(t, internalErr) {
				return
			}

			if !assert.Equal(t, 1, len(found)) {
				return
			}

			assert.True(t, areTeamsEqual(team, found[0]))
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
				if !assert.Nil(t, err) {
					return
				}
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			// create app
			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			if !assert.Nil(t, err) {
				return
			}

			defer tx.Rollback()
			app := createAppData(appID, nil, ownerUserID)
			appVersion := createAppVersionData(appID, versionNumber, true)
			team := createTeamData(teamID, ownerUserID)

			if !assert.Nil(t, appTestRef.appDaoV2.CreateApp(ct, tx, app)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionDaoV2.CreateAppVersion(ct, tx, appVersion)) {
				return
			}

			if !assert.Nil(t, appTestRef.teamDaoV2.CreateTeam(ct, tx, team)) {
				return
			}

			newAppInstallation, internalErr := appTestRef.appService.CreateAppInstallation(ct, teamID, appID,
				versionNumber)
			if testCase.expectedErr != nil {
				if !assert.NotNil(t, internalErr) {
					return
				}

				assert.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else if !assert.Nil(t, internalErr) {
				return
			}

			assert.Equal(t, appID, newAppInstallation.AppID)
			assert.Equal(t, versionNumber, newAppInstallation.EnabledVersionNumber)
			assert.Equal(t, teamID, newAppInstallation.InstalledTeamID)
			assert.Equal(t, ownerUserID, *newAppInstallation.InstalledByUserID)
			assert.NotNil(t, newAppInstallation.InstalledAt)

			// verify data in db
			appInstallationInDb, internalErr := appTestRef.appTeamInstallationDaoV2.
				FindAppTeamInstallationByAppIDAndTeamIDWithTx(ct, tx, appID, teamID)
			if !assert.Nil(t, internalErr) {
				return
			}

			assert.True(t, areAppInstallationsEqual(appInstallationInDb, newAppInstallation))
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
				if !assert.Nil(t, err) {
					return
				}
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			// create app
			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			if !assert.Nil(t, err) {
				return
			}

			defer tx.Rollback()
			app := createAppData(appID, nil, ownerUserID)
			appVersion1 := createAppVersionData(appID, versionNumber1, true)
			appVersion2 := createAppVersionData(appID, versionNumber2, true)
			team := createTeamData(teamID, ownerUserID)
			appTeamInstallation := createAppTeamInstallation(appID, versionNumber1, teamID, ownerUserID)

			if !assert.Nil(t, appTestRef.appDaoV2.CreateApp(ct, tx, app)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionDaoV2.CreateAppVersion(ct, tx, appVersion1)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionDaoV2.CreateAppVersion(ct, tx, appVersion2)) {
				return
			}

			if !assert.Nil(t, appTestRef.teamDaoV2.CreateTeam(ct, tx, team)) {
				return
			}

			if !assert.Nil(t, appTestRef.appTeamInstallationDaoV2.CreateAppTeamInstallation(ct, tx, appTeamInstallation)) {
				return
			}

			input := UpdateAppTeamInstallationInput{
				EnabledVersionNumber: versionNumber2,
			}
			updated, internalErr := appTestRef.appService.UpdateAppInstallation(ct, appID, teamID,
				input)
			if testCase.expectedErr != nil {
				if !assert.NotNil(t, internalErr) {
					return
				}

				assert.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else if !assert.Nil(t, internalErr) {
				return
			}

			appTeamInstallation.EnabledVersionNumber = versionNumber2
			assert.True(t, areAppInstallationsEqual(appTeamInstallation, updated))

			// verify data in db
			appInstallationInDb, internalErr := appTestRef.appTeamInstallationDaoV2.
				FindAppTeamInstallationByAppIDAndTeamIDWithTx(ct, tx, appID, teamID)
			if !assert.Nil(t, internalErr) {
				return
			}

			assert.True(t, areAppInstallationsEqual(appInstallationInDb, appTeamInstallation))
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
				if !assert.Nil(t, err) {
					return
				}
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			// create app
			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			if !assert.Nil(t, err) {
				return
			}

			defer tx.Rollback()
			app := createAppData(appID, nil, ownerUserID)
			appVersion := createAppVersionData(appID, versionNumber, true)
			team := createTeamData(teamID, ownerUserID)
			appTeamInstallation := createAppTeamInstallation(appID, versionNumber, teamID, ownerUserID)

			if !assert.Nil(t, appTestRef.appDaoV2.CreateApp(ct, tx, app)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionDaoV2.CreateAppVersion(ct, tx, appVersion)) {
				return
			}

			if !assert.Nil(t, appTestRef.teamDaoV2.CreateTeam(ct, tx, team)) {
				return
			}

			if !assert.Nil(t, appTestRef.appTeamInstallationDaoV2.CreateAppTeamInstallation(ct, tx, appTeamInstallation)) {
				return
			}

			deleted, internalErr := appTestRef.appService.DeleteAppInstallation(ct, appID, teamID)
			if testCase.expectedErr != nil {
				if !assert.NotNil(t, internalErr) {
					return
				}

				assert.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else if !assert.Nil(t, internalErr) {
				return
			}

			assert.True(t, areAppInstallationsEqual(appTeamInstallation, deleted))

			// verify data in db
			_, internalErr = appTestRef.appTeamInstallationDaoV2.
				FindAppTeamInstallationByAppIDAndTeamIDWithTx(ct, tx, appID, teamID)
			if !assert.NotNil(t, internalErr) {
				return
			}

			assert.Equal(t, errs.NotFound, internalErr.Code)
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
				if !assert.Nil(t, err) {
					return
				}
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			// create app
			tx, err := appTestRef.transactionFactory.BeginTx(ct, nil)
			if !assert.Nil(t, err) {
				return
			}

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

			if !assert.Nil(t, appTestRef.appDaoV2.CreateApp(ct, tx, app1)) {
				return
			}

			if !assert.Nil(t, appTestRef.appDaoV2.CreateApp(ct, tx, app2)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionDaoV2.CreateAppVersion(ct, tx, appVersion1)) {
				return
			}

			if !assert.Nil(t, appTestRef.appVersionDaoV2.CreateAppVersion(ct, tx, appVersion2)) {
				return
			}

			if !assert.Nil(t, appTestRef.teamDaoV2.CreateTeam(ct, tx, team1)) {
				return
			}

			if !assert.Nil(t, appTestRef.teamDaoV2.CreateTeam(ct, tx, team2)) {
				return
			}

			if !assert.Nil(t, appTestRef.appTeamInstallationDaoV2.CreateAppTeamInstallation(ct, tx, appTeamInstallation1)) {
				return
			}

			if !assert.Nil(t, appTestRef.appTeamInstallationDaoV2.CreateAppTeamInstallation(ct, tx, appTeamInstallation2)) {
				return
			}

			if !assert.Nil(t, appTestRef.appTeamInstallationDaoV2.CreateAppTeamInstallation(ct, tx, appTeamInstallation3)) {
				return
			}

			found, internalErr := appTestRef.appService.FindAppTeamInstallationsByAppID(ct, appID1)
			if !assert.Nil(t, internalErr) {
				return
			}

			assert.Equal(t, 2, len(found))
			for _, appTeamInstallation := range found {
				if !assert.True(t, appTeamInstallation.InstalledTeamID == teamID1 ||
					appTeamInstallation.InstalledTeamID == teamID2) {
					return
				}

				if appTeamInstallation.InstalledTeamID == teamID1 {
					assert.True(t, areAppInstallationsEqual(appTeamInstallation, appTeamInstallation1))
				} else {
					assert.True(t, areAppInstallationsEqual(appTeamInstallation, appTeamInstallation3))
				}
			}

			found, internalErr = appTestRef.appService.FindAppInstallationsByTeamID(ct, teamID2)
			if !assert.Nil(t, internalErr) {
				return
			}
			assert.Equal(t, 2, len(found))
			for _, appTeamInstallation := range found {
				if !assert.True(t, appTeamInstallation.AppID == appID1 ||
					appTeamInstallation.AppID == appID2) {
					return
				}

				if appTeamInstallation.AppID == appID1 {
					assert.True(t, areAppInstallationsEqual(appTeamInstallation, appTeamInstallation3))
				} else {
					assert.True(t, areAppInstallationsEqual(appTeamInstallation, appTeamInstallation2))
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
