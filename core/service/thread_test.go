package service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/daov2/daotestv2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type ThreadTestRef struct {
	threadService      Thread
	taskDaoV2          daov2.Task
	threadDaoV2        daov2.Thread
	messageDaoV2       daov2.Message
	transactionFactory transaction.Factory
}

func prepareThreadTestRef(t *testing.T, toggles feature.Toggles) (ThreadTestRef, bool) {
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
		return ThreadTestRef{}, false
	}

	testkit.StartServiceInstance(cloudTestKitConfig, virtualNetwork, cloudTestKit.ServiceInstanceRunner)

	teamyPrometheus := metricstest.NewNoopMetrics()
	cloudClientCfg := rpc.ConnectionConfig{
		Host:          testkit.GRPCServerHost,
		Port:          testkit.GRPCServerPort,
		ShouldEncrypt: false,
		GetAccessToken: func() string {
			return "accessToken"
		},
		RequestTimeout: 10 * time.Second,
	}
	cloudClientRegistry, err := client.NewRegistry(
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
		return ThreadTestRef{}, false
	}

	transactionFactory := transaction.NewFactory(nil)

	teamyBackendDB := dbtest.NewInMemoryDB()
	teamyBackendDB.CreateTable(daotestv2.ThreadTableName)
	teamyBackendDB.CreateTable(daotestv2.MessageTableName)
	teamyBackendDB.CreateTable(daotestv2.TaskTableName)

	teamMemberDaoV2 := daotestv2.NewTeamMember(teamyBackendDB, transactionFactory)
	stateSyncer := realtime.NewStateSyncer(logger, teamMemberDaoV2)

	taskDaoV2 := daotestv2.NewTask(teamyBackendDB, transactionFactory)
	threadDaoV2 := daotestv2.NewThread(teamyBackendDB)
	messageDaoV2 := daotestv2.NewMessage(teamyBackendDB, transactionFactory)

	threadService := NewThread(
		logger,
		toggles,
		cloudClientRegistry,
		stateSyncer,
		transactionFactory,
		taskDaoV2,
		threadDaoV2,
		messageDaoV2,
	)
	return ThreadTestRef{
		threadService: threadService,
		threadDaoV2:   threadDaoV2,
		messageDaoV2:  messageDaoV2,
		taskDaoV2:     taskDaoV2,
	}, true
}

func TestTeamService_CreateThread(t *testing.T) {
	threadRef, ok := prepareThreadTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := threadRef.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}

	defer tx.Rollback()

	// create thread
	_, err = threadRef.threadService.CreateThread(ct)
	if !assert.Nil(t, err) {
		return
	}
}

func TestTeamService_FindMessages(t *testing.T) {
	threadRef, ok := prepareThreadTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var messageID1 uint64 = 12
	var messageID2 uint64 = 13
	var threadID1 uint64 = 5
	var threadID2 uint64 = 7
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := threadRef.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}

	defer tx.Rollback()

	now := time.Now().UTC()
	message1 := entity.Message{
		ID:           messageID1,
		Body:         "Test1",
		ThreadID:     threadID1,
		AuthorUserID: requesterUserID,
		CreatedAt:    now,
		UpdatedAt:    nil,
	}

	message2 := entity.Message{
		ID:           messageID2,
		Body:         "Test2",
		ThreadID:     threadID2,
		AuthorUserID: requesterUserID,
		CreatedAt:    now,
		UpdatedAt:    nil,
	}

	// insert message into table
	if !assert.Nil(t, threadRef.messageDaoV2.CreateMessage(ct, tx, message1)) {
		return
	}

	if !assert.Nil(t, threadRef.messageDaoV2.CreateMessage(ct, tx, message2)) {
		return
	}

	messageFound, internalErr := threadRef.threadService.FindMessages(ct, threadID1)
	if !assert.Nil(t, internalErr) {
		return
	}

	// verify return result
	assert.Equal(t, 1, len(messageFound))
	assert.True(t, areMessagesEqual(message1, messageFound[0]))
}

func TestTeamService_CreateMessage(t *testing.T) {
	threadRef, ok := prepareThreadTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var threadID uint64 = 5
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := threadRef.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}

	defer tx.Rollback()

	body := "Test"
	input := CreateMessageInput{
		Body: body,
	}

	// create task
	now := time.Now().UTC()
	task := entity.Task{
		ID:               12,
		Goal:             "Test goal",
		DueAt:            &now,
		Context:          nil,
		CreatorUserID:    requesterUserID,
		OwnerUserID:      &requesterUserID,
		OwningTeamID:     5,
		Status:           entity.TaskStatusAwaiting,
		IsPlanned:        true,
		Effort:           nil,
		CommentsThreadID: threadID,
		CreatedAt:        now,
		UpdatedAt:        nil,
		DeliveredAt:      nil,
	}
	err = threadRef.taskDaoV2.CreateTask(ct, tx, task)
	if !assert.Nil(t, err) {
		return
	}

	// create message
	message, err := threadRef.threadService.CreateMessage(ct, threadID, input)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, body, message.Body)
	assert.Equal(t, threadID, message.ThreadID)

	// compare with message in memory
	messageFound, internalErr := threadRef.messageDaoV2.FindMessageByIDWithTx(ct, tx, message.ID)
	if !assert.Nil(t, internalErr) {
		return
	}

	// verify return result
	assert.True(t, areMessagesEqual(message, messageFound))
}

func TestTeamService_UpdateMessage(t *testing.T) {
	threadRef, ok := prepareThreadTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var threadID uint64 = 5
	var messageID uint64 = 12
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := threadRef.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}

	defer tx.Rollback()

	now := time.Now().UTC()
	message := entity.Message{
		ID:           messageID,
		Body:         "Test",
		ThreadID:     threadID,
		AuthorUserID: requesterUserID,
		CreatedAt:    now,
		UpdatedAt:    nil,
	}

	// create task
	task := entity.Task{
		ID:               12,
		Goal:             "Test goal",
		DueAt:            &now,
		Context:          nil,
		CreatorUserID:    requesterUserID,
		OwnerUserID:      &requesterUserID,
		OwningTeamID:     5,
		Status:           entity.TaskStatusAwaiting,
		IsPlanned:        true,
		Effort:           nil,
		CommentsThreadID: threadID,
		CreatedAt:        now,
		UpdatedAt:        nil,
		DeliveredAt:      nil,
	}
	err = threadRef.taskDaoV2.CreateTask(ct, tx, task)
	if !assert.Nil(t, err) {
		return
	}

	// insert message into table
	err = threadRef.messageDaoV2.CreateMessage(ct, tx, message)
	if !assert.Nil(t, err) {
		return
	}

	input := UpdateMessageInput{
		Body: "Updated",
	}

	// create message
	message.Body = input.Body
	updated, err := threadRef.threadService.UpdateMessage(ct, messageID, input)
	if !assert.Nil(t, err) {
		return
	}

	assert.True(t, areMessagesEqual(updated, message))

	// compare with message in memory
	messageFound, internalErr := threadRef.messageDaoV2.FindMessageByIDWithTx(ct, tx, messageID)
	if !assert.Nil(t, internalErr) {
		return
	}

	// verify return result
	assert.True(t, areMessagesEqual(messageFound, message))
}

func TestTeamService_DeleteMessage(t *testing.T) {
	threadRef, ok := prepareThreadTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var threadID uint64 = 5
	var messageID uint64 = 12
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := threadRef.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}

	defer tx.Rollback()

	now := time.Now().UTC()
	message := entity.Message{
		ID:           messageID,
		Body:         "Test",
		ThreadID:     threadID,
		AuthorUserID: requesterUserID,
		CreatedAt:    now,
		UpdatedAt:    nil,
	}

	task := entity.Task{
		ID:               12,
		Goal:             "Test goal",
		DueAt:            &now,
		Context:          nil,
		CreatorUserID:    requesterUserID,
		OwnerUserID:      &requesterUserID,
		OwningTeamID:     5,
		Status:           entity.TaskStatusAwaiting,
		IsPlanned:        true,
		Effort:           nil,
		CommentsThreadID: threadID,
		CreatedAt:        now,
		UpdatedAt:        nil,
		DeliveredAt:      nil,
	}
	err = threadRef.taskDaoV2.CreateTask(ct, tx, task)
	if !assert.Nil(t, err) {
		return
	}

	// insert message into table
	err = threadRef.messageDaoV2.CreateMessage(ct, tx, message)
	if !assert.Nil(t, err) {
		return
	}

	// create message
	deleted, err := threadRef.threadService.DeleteMessage(ct, messageID)
	if !assert.Nil(t, err) {
		return
	}

	assert.True(t, areMessagesEqual(deleted, message))

	// no message in memory
	_, internalErr := threadRef.messageDaoV2.FindMessageByIDWithTx(ct, tx, messageID)
	if !assert.NotNil(t, internalErr) {
		return
	}

	assert.Equal(t, errs.NotFound, internalErr.Code)
}

func areMessagesEqual(one entity.Message, other entity.Message) bool {
	if one.ID != other.ID {
		return false
	}

	if one.AuthorUserID != other.AuthorUserID {
		return false
	}

	if one.ThreadID != other.ThreadID {
		return false
	}

	if one.Body != other.Body {
		return false
	}

	if one.CreatedAt != other.CreatedAt {
		return false
	}

	return true
}
