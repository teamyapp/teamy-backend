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
	"github.com/teamyapp/cloud/libs/network/networktest"
	"github.com/teamyapp/cloud/libs/retry"
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/rpc"
	"github.com/teamyapp/cloud/libs/runtime"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudtx "github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/cloud/testkit"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/dao/daotest"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UserTestRef struct {
	userService              User
	userDao                  dao.User
	userFileUploadSessionDao dao.UserFileUploadSession
	transactionFactory       cloudtx.Factory
}

func prepareUserTestRef(t *testing.T, toggles feature.Toggles) (UserTestRef, bool) {
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

	testkit.StartServiceInstance(cloudTestKitConfig, virtualNetwork, cloudTestKit.ServiceInstanceRunner)

	noopMetrics := instrumenttest.NewNoopMetrics()
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

	authorizor := client.NewAuthorizer(logger, cloudClientRegistry)
	transactionFactory := cloudtx.NewFactory(nil)

	teamyBackendDB := dbtest.NewInMemoryDB()
	teamyBackendDB.CreateTable(daotest.UserTableName)
	teamyBackendDB.CreateTable(daotest.UserFileUploadSessionTableName)
	teamyBackendDB.CreateTable(daotest.TeamMemberTableName)

	teamMemberDao := daotest.NewTeamMember(teamyBackendDB, transactionFactory)
	stateSyncer := realtime.NewStateSyncer(logger, teamMemberDao)

	userDao := daotest.NewUser(teamyBackendDB, transactionFactory)
	userFileUploadSessionDao := daotest.NewUserFileUploadSession(teamyBackendDB)

	transactionGroupFactory := transaction.NewGroupFactory(logger, noopMetrics, transactionFactory, stateSyncer)
	userService := NewUser(
		logger,
		transactionGroupFactory,
		toggles,
		cloudTestKitConfig.WebAPIBaseURL,
		cloudClientRegistry,
		authorizor,
		stateSyncer,
		transactionFactory,
		userDao,
		teamMemberDao,
		userFileUploadSessionDao,
	)

	return UserTestRef{
		userService:              userService,
		userDao:                  userDao,
		userFileUploadSessionDao: userFileUploadSessionDao,
		transactionFactory:       transactionFactory,
	}, true
}

func TestUserService_CreateUser(t *testing.T) {
	userTestRef, ok := prepareUserTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
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
	newUser, internalErr := userTestRef.userService.CreateUser(ct, userInput)
	require.Nil(t, internalErr)

	require.Equal(t, requesterUserID, newUser.ID)
	require.Equal(t, userInput.LastName, newUser.LastName)
	require.Equal(t, userInput.FirstName, newUser.FirstName)
	require.Equal(t, userInput.ProfileURL, newUser.ProfileURL)
	require.NotNil(t, newUser.CreatedAt)
	require.Nil(t, newUser.UpdatedAt)

	userInMemory, err := userTestRef.userDao.FindUserByID(ct, newUser.ID)
	require.Nil(t, err)
	require.Equal(t, requesterUserID, userInMemory.ID)
	require.Equal(t, userInput.LastName, userInMemory.LastName)
	require.Equal(t, userInput.FirstName, userInMemory.FirstName)
	require.Equal(t, userInput.ProfileURL, userInMemory.ProfileURL)
	require.NotNil(t, userInMemory.CreatedAt)
	require.Nil(t, userInMemory.UpdatedAt)
}

func TestUserService_UpdateUser(t *testing.T) {
	userTestRef, ok := prepareUserTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	var profileURL = "https://test"

	tx, err := userTestRef.transactionFactory.BeginTx(ct, nil)
	require.Nil(t, err)

	defer tx.Rollback()

	user := entity.User{
		ID:         requesterUserID,
		LastName:   "LastName",
		FirstName:  "FirstName",
		ProfileURL: &profileURL,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  nil,
	}

	require.Nil(t, userTestRef.userDao.CreateUser(ct, tx, user))

	updateInput := UpdateUserInput{
		LastName:  "UpdatedLastName",
		FirstName: "UpdatedFirstName",
	}
	updatedUser, internalErr := userTestRef.userService.UpdateUser(ct, user.ID, updateInput)
	require.Nil(t, internalErr)
	require.Equal(t, requesterUserID, updatedUser.ID)
	require.Equal(t, updateInput.LastName, updatedUser.LastName)
	require.Equal(t, updateInput.FirstName, updatedUser.FirstName)
	require.Equal(t, user.ProfileURL, updatedUser.ProfileURL)
	require.Equal(t, user.CreatedAt, updatedUser.CreatedAt)
	require.NotNil(t, updatedUser.UpdatedAt)

	userInMemory, err := userTestRef.userDao.FindUserByID(ct, user.ID)
	require.Nil(t, err)
	require.Equal(t, requesterUserID, userInMemory.ID)
	require.Equal(t, updateInput.LastName, userInMemory.LastName)
	require.Equal(t, updateInput.FirstName, userInMemory.FirstName)
	require.Equal(t, user.ProfileURL, userInMemory.ProfileURL)
	require.Equal(t, user.CreatedAt, userInMemory.CreatedAt)
	require.NotNil(t, userInMemory.UpdatedAt)
}

func TestUserService_FindUserByID(t *testing.T) {
	userTestRef, ok := prepareUserTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	var profileURL = "https://test"
	tx, err := userTestRef.transactionFactory.BeginTx(ct, nil)
	require.Nil(t, err)

	defer tx.Rollback()

	now := time.Now().UTC()
	user := entity.User{
		ID:         requesterUserID,
		LastName:   "test_lastname",
		FirstName:  "test_firstname",
		ProfileURL: &profileURL,
		CreatedAt:  now,
		UpdatedAt:  &now,
	}

	require.Nil(t, userTestRef.userDao.CreateUser(ct, tx, user))

	userFound, internalErr := userTestRef.userService.FindUserByID(ct, user.ID)
	require.Nil(t, internalErr)
	require.Equal(t, user.ID, userFound.ID)
	require.Equal(t, user.LastName, userFound.LastName)
	require.Equal(t, user.FirstName, userFound.FirstName)
	require.Equal(t, user.ProfileURL, userFound.ProfileURL)
	require.Equal(t, user.CreatedAt, userFound.CreatedAt)
	require.NotNil(t, user.UpdatedAt, userFound.UpdatedAt)
}

func TestUserService_CreateUserProfileUploadSession(t *testing.T) {
	userTestRef, ok := prepareUserTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := userTestRef.transactionFactory.BeginTx(ct, nil)
	require.Nil(t, err)

	defer tx.Rollback()

	uploadSessionID, err := userTestRef.userService.CreateUserProfileUploadSession(ct)
	require.Nil(t, err)
	require.Equal(t, uploadSessionID, uint64(1))

	uploadSessionInMemory, err := userTestRef.userFileUploadSessionDao.FindUserFileUploadSessionByUserIDWithTx(ct,
		tx,
		requesterUserID,
		entity.ProfileUserFileUploadSessionType,
		uploadSessionID)
	require.Nil(t, err)
	require.Equal(t, uploadSessionInMemory.FileUploadSessionID, uploadSessionID)
	require.Equal(t, uploadSessionInMemory.UserID, requesterUserID)
	require.Nil(t, uploadSessionInMemory.UpdatedAt)
	require.Equal(t, uploadSessionInMemory.Type, entity.ProfileUserFileUploadSessionType)
	require.Equal(t, uploadSessionInMemory.IsCompleted, false)
}

func TestUserService_FinishUserProfileUploadSession(t *testing.T) {
	userTestRef, ok := prepareUserTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := userTestRef.transactionFactory.BeginTx(ct, nil)
	require.Nil(t, err)

	defer tx.Rollback()

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

	require.Nil(t, userTestRef.userDao.CreateUser(ct, tx, user))

	uploadSessionID, err := userTestRef.userService.CreateUserProfileUploadSession(ct)
	require.Nil(t, err)

	updatedUser, err := userTestRef.userService.FinishUserProfileUploadSession(ct, uploadSessionID)
	require.Nil(t, err)
	require.Equal(t, user.ID, updatedUser.ID)
	require.NotEqual(t, user.ProfileURL, updatedUser.ProfileURL)
	require.NotEqual(t, user.UpdatedAt, updatedUser.UpdatedAt)
	require.Equal(t, user.FirstName, updatedUser.FirstName)
	require.Equal(t, user.LastName, updatedUser.LastName)
	require.Equal(t, user.CreatedAt, updatedUser.CreatedAt)

	uploadSessionInMemory, err := userTestRef.userFileUploadSessionDao.FindUserFileUploadSessionByUserIDWithTx(ct,
		tx,
		requesterUserID,
		entity.ProfileUserFileUploadSessionType,
		uploadSessionID)

	require.Equal(t, uploadSessionInMemory.IsCompleted, true)
	require.Equal(t, uploadSessionInMemory.FileUploadSessionID, uploadSessionID)
	require.Equal(t, uploadSessionInMemory.UserID, user.ID)
	require.Equal(t, uploadSessionInMemory.Type, entity.ProfileUserFileUploadSessionType)
	require.NotNil(t, uploadSessionInMemory.UpdatedAt)
}
