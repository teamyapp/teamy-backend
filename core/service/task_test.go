package service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	cloudAPI "github.com/teamyapp/cloud/app/api"
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
	"github.com/teamyapp/teamy-backend/core/daov2/daotestv2"
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
					group.ID,
					teamID,
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
					group.ID,
					teamID,
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
					group.ID,
					teamID,
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
					group.ID,
					teamID,
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
				if !assert.Nil(t, err) {
					return
				}
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			now := time.Now()
			taskInput := CreateTaskInput{
				Goal:        "Unit test",
				OwnerUserID: &ownerUserID,
				IsPlanned:   true,
				DueAt:       &now,
			}
			newTask, internalErr := taskTestRef.taskService.CreateTask(ct, teamID, taskInput)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else if !assert.Nil(t, internalErr) {
				return
			}

			assert.Equal(t, uint64(1), newTask.ID)
			assert.Equal(t, taskInput.Goal, newTask.Goal)
			assert.Equal(t, taskInput.Context, newTask.Context)
			assert.Equal(t, teamID, newTask.OwningTeamID)
			assert.Equal(t, taskInput.OwnerUserID, newTask.OwnerUserID)
			assert.Equal(t, taskInput.DueAt, newTask.DueAt)
			assert.Nil(t, newTask.Effort)
			assert.Equal(t, taskInput.IsPlanned, newTask.IsPlanned)
			assert.Equal(t, entity.TaskStatusTodo, newTask.Status)
			assert.Equal(t, uint64(1), newTask.CommentsThreadID)
			assert.NotNil(t, newTask.CreatedAt)
			assert.Nil(t, newTask.UpdatedAt)
			assert.Nil(t, newTask.DeliveredAt)
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
	if !assert.Nil(t, internalErr) {
		return TaskTestRef{}, false
	}

	apiToken, internalErr := servicetest.GetServiceAccountAPIToken(cloudTestKit.IdentityService)
	if !assert.Nil(t, internalErr) {
		return TaskTestRef{}, false
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
	cloudClientRegistry, err := cloudAPI.NewClientRegistry(
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
		return TaskTestRef{}, false
	}

	authorizer := NewAuthorizer(logger, cloudClientRegistry)
	transactionFactory := transaction.NewFactory(nil)

	teamyBackendDB := dbtest.NewInMemoryDB()
	teamyBackendDB.CreateTable(daotestv2.ThreadTableName)
	teamyBackendDB.CreateTable(daotestv2.TaskTableName)

	teamMemberDaoV2 := daotestv2.NewTeamMember(teamyBackendDB, transactionFactory)
	stateSyncer := realtime.NewStateSyncer(logger, teamMemberDaoV2)
	activityCache := cache.NewActivity(logger)

	taskDaoV2 := daotestv2.NewTask(teamyBackendDB)
	threadDaoV2 := daotestv2.NewThread(teamyBackendDB)
	sprintDaoV2 := daotestv2.NewSprint(teamyBackendDB, transactionFactory)
	taskAwaitForRelationDaoV2 := daotestv2.NewTaskAwaitForRelation(teamyBackendDB)
	sprintParticipantDaoV2 := daotestv2.NewSprintParticipant(teamyBackendDB, transactionFactory)
	sprintTaskRelationDaoV2 := daotestv2.NewSprintTaskRelation(teamyBackendDB)
	taskService := NewTask(
		logger,
		cloudClientRegistry,
		authorizer,
		toggles,
		stateSyncer,
		transactionFactory,
		activityCache,
		taskDaoV2,
		threadDaoV2,
		sprintDaoV2,
		taskAwaitForRelationDaoV2,
		sprintParticipantDaoV2,
		sprintTaskRelationDaoV2,
	)
	return TaskTestRef{
		taskService:  taskService,
		cloudTestKit: cloudTestKit,
	}, true
}
