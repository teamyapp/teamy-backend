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
	"github.com/teamyapp/cloud/libs/metrics/metricstest"
	"github.com/teamyapp/cloud/libs/network/networktest"
	"github.com/teamyapp/cloud/libs/retry"
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/rpc"
	"github.com/teamyapp/cloud/libs/runtime"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/cloud/testkit"
	"github.com/teamyapp/teamy-backend/core/dao/daotest"
	"github.com/teamyapp/teamy-backend/core/daov2/daotestv2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

func prepareUserService(t *testing.T) (User, bool) {
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
		SlackClientSecret:        "SlackSecret",
		WebServerPort:            80,
		GRPCServerPort:           81,
	}
	cloudTestKit, internalErr := testkit.New(cloudTestKitConfig, virtualNetwork)
	if !assert.Nil(t, internalErr) {
		return User{}, false
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
	cloudClientRegistry, err := cloudAPI.NewClientRegistry(
		dataCollector,
		virtualNetwork,
		teamyPrometheus,
		cloudClientCfg,
		func() retry.Retry {
			exponentialBackOff := backoff.NewExponentialBuilder().Build()
			return retry.NewMaxCount(runtime.NewBuiltInRuntime(), exponentialBackOff, 3)
		})
	if !assert.Nil(t, err) {
		return User{}, false
	}

	teamyBackendDB := dbtest.NewInMemoryDB()
	teamyBackendDB.CreateTable(daotestv2.UserTableName)
	teamyBackendDB.CreateTable(daotestv2.UserFileUploadSessionTableName)
	teamyBackendDB.CreateTable(daotestv2.TeamMemberTableName)

	teamMemberDao := daotest.NewTeamMember(teamyBackendDB)
	stateSyncer := realtime.NewStateSyncer(dataCollector, teamMemberDao)
	transactionFactory := transaction.NewFactory(nil)

	userDao := daotest.NewUser(teamyBackendDB)
	userDaoV2 := daotestv2.NewUser(teamyBackendDB)
	teamMemberDaoV2 := daotestv2.NewTeamMember(teamyBackendDB)
	userFileUploadSessionDaoV2 := daotestv2.NewUserFileUploadSession(teamyBackendDB)
	return NewUser(
		dataCollector,
		cloudTestKitConfig.WebAPIBaseURL,
		cloudClientRegistry,
		stateSyncer,
		transactionFactory,
		userDao,
		userDaoV2,
		teamMemberDaoV2,
		userFileUploadSessionDaoV2,
	), true
}

func TestUserService_CreateUser(t *testing.T) {
	userService, ok := prepareUserService(t)
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	var profileURL = "https://test"
	userInput := CreateUserInput{
		LastName:   "LastName",
		FirstName:  "FirstName",
		ProfileURL: &profileURL,
	}
	newUser, internalErr := userService.CreateUser(ct, userInput)
	if !assert.Nil(t, internalErr) {
		return
	}

	// verify return result
	assert.Equal(t, requesterUserID, newUser.ID)
	assert.Equal(t, userInput.LastName, newUser.LastName)
	assert.Equal(t, userInput.FirstName, newUser.FirstName)
	assert.Equal(t, userInput.ProfileURL, newUser.ProfileURL)
	assert.NotNil(t, newUser.CreatedAt)
	assert.Nil(t, newUser.UpdatedAt)

	// verify in-memory DB
	userInMemory, err := userService.userDaoV2.FindUserByID(ct, newUser.ID)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, requesterUserID, userInMemory.ID)
	assert.Equal(t, userInput.LastName, userInMemory.LastName)
	assert.Equal(t, userInput.FirstName, userInMemory.FirstName)
	assert.Equal(t, userInput.ProfileURL, userInMemory.ProfileURL)
	assert.NotNil(t, userInMemory.CreatedAt)
	assert.Nil(t, userInMemory.UpdatedAt)
}

func TestUserService_UpdateUser(t *testing.T) {
	userService, ok := prepareUserService(t)
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	var profileURL = "https://test"

	tx, err := userService.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}

	user := entity.User{
		ID:         requesterUserID,
		LastName:   "LastName",
		FirstName:  "FirstName",
		ProfileURL: &profileURL,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  nil,
	}

	// insert user into table
	if !assert.Nil(t, userService.userDaoV2.CreateUser(ct, tx, user)) {
		return
	}

	updateInput := UpdateUserInput{
		LastName:  "UpdatedLastName",
		FirstName: "UpdatedFirstName",
	}
	updatedUser, internalErr := userService.UpdateUser(ct, user.ID, updateInput)
	if !assert.Nil(t, internalErr) {
		return
	}

	// verify return result
	assert.Equal(t, requesterUserID, updatedUser.ID)
	assert.Equal(t, updateInput.LastName, updatedUser.LastName)
	assert.Equal(t, updateInput.FirstName, updatedUser.FirstName)
	assert.Equal(t, user.ProfileURL, updatedUser.ProfileURL)
	assert.Equal(t, user.CreatedAt, updatedUser.CreatedAt)
	assert.NotNil(t, updatedUser.UpdatedAt)

	// verify in-memory DB
	userInMemory, err := userService.userDaoV2.FindUserByID(ct, user.ID)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, requesterUserID, userInMemory.ID)
	assert.Equal(t, updateInput.LastName, userInMemory.LastName)
	assert.Equal(t, updateInput.FirstName, userInMemory.FirstName)
	assert.Equal(t, user.ProfileURL, userInMemory.ProfileURL)
	assert.Equal(t, user.CreatedAt, userInMemory.CreatedAt)
	assert.NotNil(t, userInMemory.UpdatedAt)
}

func TestUserService_FindUserByID(t *testing.T) {
	userService, ok := prepareUserService(t)
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	var profileURL = "https://test"
	tx, err := userService.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}

	now := time.Now().UTC()
	user := entity.User{
		ID:         requesterUserID,
		LastName:   "test_lastname",
		FirstName:  "test_firstname",
		ProfileURL: &profileURL,
		CreatedAt:  now,
		UpdatedAt:  &now,
	}

	// insert user into table
	if !assert.Nil(t, userService.userDaoV2.CreateUser(ct, tx, user)) {
		return
	}

	userFound, internalErr := userService.FindUserByID(ct, user.ID)
	if !assert.Nil(t, internalErr) {
		return
	}

	// verify return result
	assert.Equal(t, user.ID, userFound.ID)
	assert.Equal(t, user.LastName, userFound.LastName)
	assert.Equal(t, user.FirstName, userFound.FirstName)
	assert.Equal(t, user.ProfileURL, userFound.ProfileURL)
	assert.Equal(t, user.CreatedAt, userFound.CreatedAt)
	assert.NotNil(t, user.UpdatedAt, userFound.UpdatedAt)
}

func TestUserService_CreateUserProfileUploadSession(t *testing.T) {
	userService, ok := prepareUserService(t)
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := userService.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}

	uploadSessionID, err := userService.CreateUserProfileUploadSession(ct)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, uploadSessionID, uint64(1))

	// verify in-memory DB
	uploadSessionInMemory, err := userService.userFileUploadSessionDaoV2.FindUserFileUploadSessionByUserIDWithTx(ct,
		tx,
		requesterUserID,
		entity.ProfileUserFileUploadSessionType,
		uploadSessionID)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, uploadSessionInMemory.FileUploadSessionID, uploadSessionID)
	assert.Equal(t, uploadSessionInMemory.UserID, requesterUserID)
	assert.Nil(t, uploadSessionInMemory.UpdatedAt)
	assert.Equal(t, uploadSessionInMemory.Type, entity.ProfileUserFileUploadSessionType)
	assert.Equal(t, uploadSessionInMemory.IsCompleted, false)
}

func TestUserService_UpdateUserProfileUploadSession(t *testing.T) {
	userService, ok := prepareUserService(t)
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := userService.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}

	var profileURL = "https://test"
	now := time.Now().UTC()
	user := entity.User{
		ID:         requesterUserID,
		LastName:   "test_lastname",
		FirstName:  "test_firstname",
		ProfileURL: &profileURL,
		CreatedAt:  now,
		UpdatedAt:  &now,
	}

	// insert user table
	if !assert.Nil(t, userService.userDaoV2.CreateUser(ct, tx, user)) {
		return
	}

	// create user upload session
	uploadSessionID, err := userService.CreateUserProfileUploadSession(ct)
	if !assert.Nil(t, err) {
		return
	}

	// finish upload session
	updatedUser, err := userService.FinishUserProfileUploadSession(ct, uploadSessionID)
	if !assert.Nil(t, err) {
		return
	}

	// verify returned user
	assert.Equal(t, user.ID, updatedUser.ID)
	assert.NotEqual(t, user.ProfileURL, updatedUser.ProfileURL)
	assert.NotEqual(t, user.UpdatedAt, updatedUser.UpdatedAt)
	assert.Equal(t, user.FirstName, updatedUser.FirstName)
	assert.Equal(t, user.LastName, updatedUser.LastName)
	assert.Equal(t, user.CreatedAt, updatedUser.CreatedAt)

	// verify in-memory DB
	uploadSessionInMemory, err := userService.userFileUploadSessionDaoV2.FindUserFileUploadSessionByUserIDWithTx(ct,
		tx,
		requesterUserID,
		entity.ProfileUserFileUploadSessionType,
		uploadSessionID)

	assert.Equal(t, uploadSessionInMemory.IsCompleted, true)
	assert.Equal(t, uploadSessionInMemory.FileUploadSessionID, uploadSessionID)
	assert.Equal(t, uploadSessionInMemory.UserID, user.ID)
	assert.Equal(t, uploadSessionInMemory.Type, entity.ProfileUserFileUploadSessionType)
	assert.NotNil(t, uploadSessionInMemory.UpdatedAt)
}
