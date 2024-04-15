package service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/teamyapp/teamy-backend/core/instrument/instrumenttest"
	"github.com/teamyapp/teamy-backend/core/transaction"

	"github.com/stretchr/testify/require"
	"github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/network/networktest"
	"github.com/teamyapp/cloud/libs/retry"
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/rpc"
	"github.com/teamyapp/cloud/libs/runtime"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudtx "github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/cloud/testkit"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/cache"
	"github.com/teamyapp/teamy-backend/core/dao/daotest"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/service/servicetest"
)

type TaskTestRef struct {
	taskService  Task
	cloudTestKit testkit.TestKit
}

func TestTaskService_CreateTask(t *testing.T) {
	var teamID uint64 = 1
	var ownerUserID uint64 = 3
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(taskTestRef TaskTestRef, requesterUserID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed when user is team owner",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(taskTestRef TaskTestRef, requesterUserID uint64) *errs.Error {
				ct := context.Background()
				ct = ctx.NewContextWithUserID(ct, 1)
				group, err := taskTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(ct, "Owner", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					ct,
					taskTestRef.cloudTestKit.AuthorizationService,
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
				EnableAuthorization: true,
			},
			prepareData: func(taskTestRef TaskTestRef, requesterUserID uint64) *errs.Error {
				ct := context.Background()
				ct = ctx.NewContextWithUserID(ct, 1)
				group, err := taskTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(ct, "Admin", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					ct,
					taskTestRef.cloudTestKit.AuthorizationService,
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
				EnableAuthorization: true,
			},
			prepareData: func(taskTestRef TaskTestRef, requesterUserID uint64) *errs.Error {
				ct := context.Background()
				ct = ctx.NewContextWithUserID(ct, 1)
				group, err := taskTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(ct, "Member", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					ct,
					taskTestRef.cloudTestKit.AuthorizationService,
					teamID,
					group.ID,
					authorization.TeamMemberResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID: 3,
			expectedErr:     nil,
		},
		{
			name: "permission denied when user is not in team",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(taskTestRef TaskTestRef, requesterUserID uint64) *errs.Error {
				ct := context.Background()
				ct = ctx.NewContextWithUserID(ct, 1)
				group, err := taskTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(ct, "Member", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					ct,
					taskTestRef.cloudTestKit.AuthorizationService,
					teamID,
					group.ID,
					authorization.TeamMemberResourceTypeOperations,
					3)
			},
			requesterUserID: 4,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			taskTestRef, ok := prepareTaskTestRef(t, testCase.toggles)
			if !ok {
				return
			}

			if testCase.prepareData != nil {
				err := testCase.prepareData(taskTestRef, testCase.requesterUserID)
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			now := time.Now()
			taskInput := CreateTaskInput{
				Goal:        "Unit test",
				OwnerUserID: &ownerUserID,
				IsScheduled: true,
				DueAt:       &now,
			}
			newTask, internalErr := taskTestRef.taskService.CreateTask(ct, teamID, taskInput)
			if testCase.expectedErr != nil {
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			}

			require.Nil(t, internalErr)
			require.Equal(t, uint64(1), newTask.ID)
			require.Equal(t, taskInput.Goal, newTask.Goal)
			require.Equal(t, taskInput.Context, newTask.Context)
			require.Equal(t, teamID, newTask.OwningTeamID)
			require.Equal(t, taskInput.OwnerUserID, newTask.OwnerUserID)
			require.Equal(t, taskInput.DueAt, newTask.DueAt)
			require.Nil(t, newTask.Effort)
			require.Equal(t, taskInput.IsScheduled, newTask.IsScheduled)
			require.Equal(t, entity.TaskStatusTodo, newTask.Status)
			require.Equal(t, uint64(1), newTask.CommentsThreadID)
			require.NotNil(t, newTask.CreatedAt)
			require.Nil(t, newTask.UpdatedAt)
			require.Nil(t, newTask.DeliveredAt)
		})
	}
}

func prepareTaskTestRef(t *testing.T, toggles feature.Toggles) (TaskTestRef, bool) {
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

	apiToken, internalErr := servicetest.GetServiceAccountAPIToken(cloudTestKit.IdentityService)
	require.Nil(t, internalErr)

	testkit.StartServiceInstance(cloudTestKitConfig, virtualNetwork, cloudTestKit.ServiceInstanceRunner)
	noopMetrics := instrumenttest.NewNoopMetrics()
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
	transactionFactory := cloudtx.NewFactory(nil)

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
	storyTaskRelationDao := daotest.NewStoryTaskRelation(teamyBackendDB)
	transactionGroupFactory := transaction.NewGroupFactory(logger, noopMetrics, transactionFactory, stateSyncer)
	taskService := NewTask(
		logger,
		transactionGroupFactory,
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
		storyTaskRelationDao,
	)
	return TaskTestRef{
		taskService:  taskService,
		cloudTestKit: cloudTestKit,
	}, true
}
