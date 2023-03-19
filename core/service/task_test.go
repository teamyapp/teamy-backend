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
	"github.com/teamyapp/cloud/libs/env"
	"github.com/teamyapp/cloud/libs/metrics"
	"github.com/teamyapp/cloud/libs/network/networktest"
	"github.com/teamyapp/cloud/libs/retry"
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/rpc"
	"github.com/teamyapp/cloud/libs/runtime"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/cloud/testkit"
	"github.com/teamyapp/teamy-backend/core/cache"
	"github.com/teamyapp/teamy-backend/core/dao/daotest"
	"github.com/teamyapp/teamy-backend/core/daov2/daotestv2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

func TestTaskService_CreateTask(t *testing.T) {
	lineFormatter := telemetry.NewOrderedColumnLineFormatter([]string{})
	logger := telemetry.NewLogger(lineFormatter, os.Stdout, telemetry.Off, []telemetry.LogInterceptor{})
	dataCollector := telemetry.NewDataCollector(logger)
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
	assert.Nil(t, internalErr)
	if internalErr != nil {
		return
	}

	testkit.StartServiceInstance(cloudTestKitConfig, virtualNetwork, cloudTestKit.ServiceInstanceRunner)

	teamyPrometheus := metrics.NewPrometheus("teamy", "backend", env.DevelopmentEnv)
	cloudClientCfg := rpc.ConnectionConfig{
		Host:          testkit.GRPCServerHost,
		Port:          testkit.GRPCServerPort,
		ShouldEncrypt: false,
		GetAccessToken: func() string {
			return "accessToken"
		},
		RequestTimeout: 10 * time.Second,
	}
	cloudClientRegistry, err := cloudAPI.NewClientRegistry(
		dataCollector,
		virtualNetwork,
		teamyPrometheus,
		cloudClientCfg,
		func() retry.Retry {
			exponentialBackOff := backoff.NewExponentialBuilder().Build()
			return retry.NewMaxCount(runtime.NewBuiltInRuntime(), exponentialBackOff, 3)
		})
	assert.Nil(t, err)
	if err != nil {
		return
	}

	authorizer := NewAuthorizer(dataCollector, cloudClientRegistry)

	teamyBackendDB := dbtest.NewInMemoryDB()
	teamyBackendDB.CreateTable(daotestv2.ThreadTableName)
	teamyBackendDB.CreateTable(daotestv2.TaskTableName)

	teamMemberDao := daotest.NewTeamMember(teamyBackendDB)
	stateSyncer := realtime.NewStateSyncer(dataCollector, teamMemberDao)
	transactionFactory := transaction.NewFactory(nil)
	activityCache := cache.NewActivity(dataCollector)

	taskDao := daotest.NewTask(teamyBackendDB)
	taskDaoV2 := daotestv2.NewTask(teamyBackendDB)
	threadDaoV2 := daotestv2.NewThread(teamyBackendDB)
	sprintDao := daotest.NewSprint(teamyBackendDB)
	sprintDaoV2 := daotestv2.NewSprint(teamyBackendDB)
	taskAwaitForRelationDao := daotest.NewTaskAwaitForRelation(teamyBackendDB)
	taskAwaitForRelationDaoV2 := daotestv2.NewTaskAwaitForRelation(teamyBackendDB)
	sprintParticipantDao := daotest.NewSprintParticipant(teamyBackendDB)
	sprintParticipantDaoV2 := daotestv2.NewSprintParticipant(teamyBackendDB)
	sprintTaskRelationDao := daotest.NewSprintTaskRelation(teamyBackendDB)
	sprintTaskRelationDaoV2 := daotestv2.NewSprintTaskRelation(teamyBackendDB)
	taskService := NewTask(
		dataCollector,
		cloudClientRegistry,
		authorizer,
		stateSyncer,
		transactionFactory,
		activityCache,
		taskDao,
		taskDaoV2,
		threadDaoV2,
		sprintDao,
		sprintDaoV2,
		taskAwaitForRelationDao,
		taskAwaitForRelationDaoV2,
		sprintParticipantDao,
		sprintParticipantDaoV2,
		sprintTaskRelationDao,
		sprintTaskRelationDaoV2,
	)

	var requesterUserID uint64 = 2
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	var teamID uint64 = 1
	var ownerUserID uint64 = 2
	now := time.Now()
	taskInput := CreateTaskInput{
		Goal:        "Unit test",
		OwnerUserID: &ownerUserID,
		IsPlanned:   true,
		DueAt:       &now,
	}
	newTask, internalErr := taskService.CreateTask(ct, teamID, taskInput)
	assert.Nil(t, internalErr)
	if internalErr != nil {
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
}
