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
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type TaskLinkTestRef struct {
	taskService     Task
	taskLinkService TaskLink
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
	if !assert.Nil(t, internalErr) {
		return TaskLinkTestRef{}, false
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
		return TaskLinkTestRef{}, false
	}

	authorizer := NewAuthorizer(logger, cloudClientRegistry)
	transactionFactory := transaction.NewFactory(nil)

	teamyBackendDB := dbtest.NewInMemoryDB()

	teamyBackendDB.CreateTable(daotestv2.ThreadTableName)
	teamyBackendDB.CreateTable(daotestv2.TaskTableName)

	teamMemberDao := daotest.NewTeamMember(teamyBackendDB)
	teamMemberDaov2 := daotestv2.NewTeamMember(teamyBackendDB, transactionFactory)
	stateSyncer := realtime.NewStateSyncer(logger, teamMemberDao, teamMemberDaov2)
	activityCache := cache.NewActivity(logger)

	taskDao := daotest.NewTask(teamyBackendDB)
	taskDaoV2 := daotestv2.NewTask(teamyBackendDB)
	threadDaoV2 := daotestv2.NewThread(teamyBackendDB)
	sprintDao := daotest.NewSprint(teamyBackendDB)
	sprintDaoV2 := daotestv2.NewSprint(teamyBackendDB, transactionFactory)
	taskAwaitForRelationDao := daotest.NewTaskAwaitForRelation(teamyBackendDB)
	taskAwaitForRelationDaoV2 := daotestv2.NewTaskAwaitForRelation(teamyBackendDB)
	sprintParticipantDao := daotest.NewSprintParticipant(teamyBackendDB)
	sprintParticipantDaoV2 := daotestv2.NewSprintParticipant(teamyBackendDB, transactionFactory)
	sprintTaskRelationDao := daotest.NewSprintTaskRelation(teamyBackendDB)
	sprintTaskRelationDaoV2 := daotestv2.NewSprintTaskRelation(teamyBackendDB)
	taskService := NewTask(
		logger,
		cloudClientRegistry,
		authorizer,
		toggles,
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

	teamyBackendDB.CreateTable(daotestv2.TaskLinkTableName)
	taskLinkDaoV2 := daotestv2.NewTaskLink(teamyBackendDB)

	toggles = feature.Toggles{
		EnableAuthorization: false,
	}
	taskLinkService := NewTaskLink(
		logger,
		cloudClientRegistry,
		transactionFactory,
		authorizer,
		toggles,
		stateSyncer,
		taskLinkDaoV2,
		taskDaoV2,
	)
	return TaskLinkTestRef{
		taskService:     taskService,
		taskLinkService: taskLinkService,
	}, true
}

func TestTaskLinkService_CreateTaskLink(t *testing.T) {
	taskLinkTestRef, ok := prepareTaskLinkTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

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
	newTask, internalErr := taskLinkTestRef.taskService.CreateTask(ct, teamID, taskInput)

	if !assert.Nil(t, internalErr) {
		return
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
	if !assert.Nil(t, internalErr) {
		return
	}

	assert.Equal(t, uint64(1), newTaskLink.ID)
	assert.Equal(t, taskLinkInput.TaskID, newTaskLink.TaskID)
	assert.Equal(t, taskLinkInput.Title, newTaskLink.Title)
	assert.Equal(t, taskLinkInput.URL, newTaskLink.URL)
	assert.Equal(t, taskLinkInput.IconURL, newTaskLink.IconURL)
	assert.Equal(t, taskLinkInput.IconHoverURL, newTaskLink.IconHoverURL)
	assert.NotNil(t, newTaskLink.CreatedAt)
	assert.Nil(t, newTaskLink.UpdatedAt)
}
