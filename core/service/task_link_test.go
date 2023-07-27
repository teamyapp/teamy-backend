package service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/teamyapp/cloud/app/client"
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
	"github.com/teamyapp/teamy-backend/core/cache"
	"github.com/teamyapp/teamy-backend/core/dao/daotest"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/service/servicetest"
)

type TaskLinkTestRef struct {
	teamService     Team
	taskService     Task
	taskLinkService TaskLink
	cloudTestKit    testkit.TestKit
}

func prepareTaskLinkTestRef(t *testing.T, toggles feature.Toggles) (TaskLinkTestRef, bool) {
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
		SlackClientSecret:        "Slack",
		WebServerPort:            80,
		GRPCServerPort:           81,
	}
	cloudTestKit, internalErr := testkit.New(cloudTestKitConfig, virtualNetwork)
	require.Nil(t, internalErr)

	testkit.StartServiceInstance(cloudTestKitConfig, virtualNetwork, cloudTestKit.ServiceInstanceRunner)
	apiToken, internalErr := servicetest.GetServiceAccountAPIToken(cloudTestKit.IdentityService)
	require.Nil(t, internalErr)

	noopMetrics := metricstest.NewNoopMetrics()
	cloudClientCfg := rpc.ConnectionConfig{
		Host:          testkit.GRPCServerHost,
		Port:          testkit.GRPCServerPort,
		ShouldEncrypt: false,
		GetAccessToken: func() string {
			return apiToken
		},
		RequestTimeout: 10 * time.Second,
	}
	cloudClientRegistry, err := client.NewRegistry(
		logger,
		virtualNetwork,
		noopMetrics,
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

	authorizer := client.NewAuthorizer(logger, cloudClientRegistry)
	transactionFactory := transaction.NewFactory(nil)

	teamyBackendDB := dbtest.NewInMemoryDB()
	teamyBackendDB.CreateTable(daotest.ThreadTableName)
	teamyBackendDB.CreateTable(daotest.TaskTableName)

	teamMemberDao := daotest.NewTeamMember(teamyBackendDB, transactionFactory)
	stateSyncer := realtime.NewStateSyncer(logger, teamMemberDao)
	activityCache := cache.NewActivity(logger)

	taskDao := daotest.NewTask(teamyBackendDB, transactionFactory)
	threadDao := daotest.NewThread(teamyBackendDB)
	sprintDao := daotest.NewSprint(teamyBackendDB, transactionFactory)
	taskAwaitForRelationDao := daotest.NewTaskAwaitForRelation(teamyBackendDB)
	sprintParticipantDao := daotest.NewSprintParticipant(teamyBackendDB, transactionFactory)
	sprintTaskRelationDao := daotest.NewSprintTaskRelation(teamyBackendDB)
	teamDao := daotest.NewTeam(teamyBackendDB, transactionFactory)
	teamFileUploadSessionDao := daotest.NewTeamFileUploadSession(teamyBackendDB)
	teamGroupDao := daotest.NewTeamGroup(teamyBackendDB, transactionFactory)
	teamService := NewTeam(
		logger,
		cloudTestKitConfig.WebAPIBaseURL,
		cloudClientRegistry,
		authorizer,
		toggles,
		stateSyncer,
		transactionFactory,
		taskDao,
		sprintDao,
		sprintParticipantDao,
		teamDao,
		teamMemberDao,
		teamFileUploadSessionDao,
		teamGroupDao,
	)
	taskService := NewTask(
		logger,
		cloudClientRegistry,
		authorizer,
		toggles,
		stateSyncer,
		transactionFactory,
		activityCache,
		taskDao,
		threadDao,
		sprintDao,
		taskAwaitForRelationDao,
		sprintParticipantDao,
		sprintTaskRelationDao,
	)

	teamyBackendDB.CreateTable(daotest.TaskLinkTableName)
	taskLinkDao := daotest.NewTaskLink(teamyBackendDB)
	taskLinkService := NewTaskLink(
		logger,
		cloudClientRegistry,
		transactionFactory,
		authorizer,
		toggles,
		stateSyncer,
		taskLinkDao,
		taskDao,
	)
	return TaskLinkTestRef{
		teamService:     teamService,
		taskService:     taskService,
		taskLinkService: taskLinkService,
		cloudTestKit:    cloudTestKit,
	}, true
}

func TestTaskLinkService_CreateTaskLink(t *testing.T) {
	var teamID uint64 = 1
	var ownerUserID uint64 = 3
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(taskLinkTestRef TaskLinkTestRef, requesterUserID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed when user is team owner",
			toggles: feature.Toggles{
				EnableAuthorization: false,
			},
			prepareData: func(taskLinkTestRef TaskLinkTestRef, requesterUserID uint64) *errs.Error {
				ct := context.Background()
				ct = ctx.NewContextWithUserID(ct, 1)
				group, err := taskLinkTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(ct, "Owner", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					ct,
					taskLinkTestRef.cloudTestKit.AuthorizationService,
					teamID,
					group.ID,
					authorization.TeamOwnerResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "succeed when user is team admin",
			toggles: feature.Toggles{
				EnableAuthorization: false,
			},
			prepareData: func(taskLinkTestRef TaskLinkTestRef, requesterUserID uint64) *errs.Error {
				ct := context.Background()
				ct = ctx.NewContextWithUserID(ct, 1)
				group, err := taskLinkTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(ct, "Admin", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					ct,
					taskLinkTestRef.cloudTestKit.AuthorizationService,
					teamID,
					group.ID,
					authorization.TeamAdminResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "succeed when user is team member",
			toggles: feature.Toggles{
				EnableAuthorization: false,
			},
			prepareData: func(taskLinkTestRef TaskLinkTestRef, requesterUserID uint64) *errs.Error {
				ct := context.Background()
				ct = ctx.NewContextWithUserID(ct, 1)
				group, err := taskLinkTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(ct, "Member", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					ct,
					taskLinkTestRef.cloudTestKit.AuthorizationService,
					teamID,
					group.ID,
					authorization.TeamAdminResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID: 3,
			expectedErr:     nil,
		},
		//{
		//	name: "permission denied when user is not in team",
		//	toggles: feature.Toggles{
		//		EnableAuthorization: true,
		//	},
		//	prepareData: func(taskLinkTestRef TaskLinkTestRef, requesterUserID uint64) *errs.Error {
		//		ct := context.Background()
		//		ct = ctx.NewContextWithUserID(ct, 1)
		//		group, err := taskLinkTestRef.
		//			cloudTestKit.
		//			AuthorizationService.
		//			CreateUserGroup(ct, "Member", nil)
		//		if err != nil {
		//			return err
		//		}
		//
		//		return servicetest.AddTeamPermission(
		//			ct,
		//			taskLinkTestRef.cloudTestKit.AuthorizationService,
		//			group.ID,
		//			teamID,
		//			authorization.TeamMemberResourceTypeOperations,
		//			3)
		//	},
		//	requesterUserID: 4,
		//	expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		//},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			taskLinkTestRef, ok := prepareTaskLinkTestRef(t, testCase.toggles)
			if !ok {
				return
			}

			if testCase.prepareData != nil {
				err := testCase.prepareData(taskLinkTestRef, testCase.requesterUserID)
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			now := time.Now().UTC()
			taskInput := CreateTaskInput{
				Goal:        "Unit test",
				OwnerUserID: &ownerUserID,
				IsPlanned:   true,
				DueAt:       &now,
			}
			newTask, internalErr := taskLinkTestRef.taskService.CreateTask(ct, teamID, taskInput)
			if testCase.expectedErr != nil {
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			IconURL := "task link icon url"
			IconHoverURL := "task link hover url"
			taskLinkInput := CreateTaskLinkInput{
				TaskID:       newTask.ID,
				Title:        "task link title",
				URL:          "task link url",
				IconURL:      &IconURL,
				IconHoverURL: &IconHoverURL,
			}
			newTaskLink, internalErr := taskLinkTestRef.taskLinkService.CreateTaskLink(ct, taskLinkInput)
			require.Nil(t, internalErr)

			require.Equal(t, uint64(1), newTaskLink.ID)
			require.Equal(t, taskLinkInput.TaskID, newTaskLink.TaskID)
			require.Equal(t, taskLinkInput.Title, newTaskLink.Title)
			require.Equal(t, taskLinkInput.URL, newTaskLink.URL)
			require.Equal(t, taskLinkInput.IconURL, newTaskLink.IconURL)
			require.Equal(t, taskLinkInput.IconHoverURL, newTaskLink.IconHoverURL)
			require.NotNil(t, newTaskLink.CreatedAt)
			require.Nil(t, newTaskLink.UpdatedAt)
		})
	}
}
