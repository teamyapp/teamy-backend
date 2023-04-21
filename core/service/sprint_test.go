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
	"github.com/teamyapp/teamy-backend/core/dao/daotest"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/daov2/daotestv2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/service/servicetest"
)

type SprintTestRef struct {
	sprintService      Sprint
	sprintDaoV2        daov2.Sprint
	teamDaoV2          daov2.Team
	userDaoV2          daov2.User
	cloudTestKit       testkit.TestKit
	transactionFactory transaction.Factory
}

func prepareSprintTestRef(t *testing.T, toggles feature.Toggles) (SprintTestRef, bool) {
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
		return SprintTestRef{}, false
	}

	testkit.StartServiceInstance(cloudTestKitConfig, virtualNetwork, cloudTestKit.ServiceInstanceRunner)
	apiToken, internalErr := servicetest.GetServiceAccountAPIToken(cloudTestKit.IdentityService)
	if !assert.Nil(t, internalErr) {
		return SprintTestRef{}, false
	}

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
		return SprintTestRef{}, false
	}

	authorizer := NewAuthorizer(logger, cloudClientRegistry)

	transactionFactory := transaction.NewFactory(nil)
	teamyBackendDB := dbtest.NewInMemoryDB()
	teamyBackendDB.CreateTable(daotestv2.TeamTableName)
	teamyBackendDB.CreateTable(daotestv2.SprintTableName)
	teamyBackendDB.CreateTable(daotestv2.TeamMemberTableName)
	// teamyBackendDB.CreateTable(daotestv2.UserTableName)
	teamMemberDao := daotest.NewTeamMember(teamyBackendDB)
	teamMemberDaoV2 := daotestv2.NewTeamMember(teamyBackendDB, transactionFactory)
	taskDao := daotest.NewTask(teamyBackendDB)
	sprintDao := daotest.NewSprint(teamyBackendDB)
	sprintDaoV2 := daotestv2.NewSprint(teamyBackendDB, transactionFactory)
	taskDaoV2 := daotestv2.NewTask(teamyBackendDB)
	stateSyncer := realtime.NewStateSyncer(logger, teamMemberDao, teamMemberDaoV2)
	teamDaoV2 := daotestv2.NewTeam(teamyBackendDB, transactionFactory)
	sprintTaskRelationDao := daotest.NewSprintTaskRelation(teamyBackendDB)
	sprintTaskRelationDaoV2 := daotestv2.NewSprintTaskRelation(teamyBackendDB)
	sprintParticipantDao := daotest.NewSprintParticipant(teamyBackendDB)
	sprintParticipantDaoV2 := daotestv2.NewSprintParticipant(teamyBackendDB, transactionFactory)
	threadDaoV2 := daotestv2.NewThread(teamyBackendDB)
	userDaoV2 := daotestv2.NewUser(teamyBackendDB, transactionFactory)

	sprintService := NewSprint(
		logger,
		cloudClientRegistry,
		stateSyncer,
		authorizer,
		toggles,
		transactionFactory,
		taskDao,
		taskDaoV2,
		sprintDao,
		sprintDaoV2,
		sprintTaskRelationDao,
		sprintTaskRelationDaoV2,
		sprintParticipantDao,
		sprintParticipantDaoV2,
		teamMemberDao,
		teamMemberDaoV2,
		threadDaoV2,
	)

	return SprintTestRef{
		sprintService:      sprintService,
		cloudTestKit:       cloudTestKit,
		sprintDaoV2:        sprintDaoV2,
		teamDaoV2:          teamDaoV2,
		userDaoV2:          userDaoV2,
		transactionFactory: transactionFactory,
	}, true
}

func TestSprintService_CreateSprint(t *testing.T) {
	var teamID uint64 = 1
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(sprintTestRef SprintTestRef, requesterUserID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed when user is team owner",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(sprintTestRef SprintTestRef, requesterUserID uint64) *errs.Error {
				ct := context.Background()
				ct = ctx.NewContextWithUserID(ct, 1)
				group, err := sprintTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(ct, "Owner", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					ct,
					sprintTestRef.cloudTestKit.AuthorizationService,
					group.ID,
					teamID,
					authorization.TeamOwnerResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID: 1,
			expectedErr:     nil,
		},
		{
			name: "succeed when user is team admin",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(sprintTestRef SprintTestRef, requesterUserID uint64) *errs.Error {
				ct := context.Background()
				ct = ctx.NewContextWithUserID(ct, 1)
				group, err := sprintTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(ct, "Admin", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					ct,
					sprintTestRef.cloudTestKit.AuthorizationService,
					group.ID,
					teamID,
					authorization.TeamAdminResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID: 1,
			expectedErr:     nil,
		},
		{
			name: "succeed when user is team member",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(sprintTestRef SprintTestRef, requesterUserID uint64) *errs.Error {
				ct := context.Background()
				ct = ctx.NewContextWithUserID(ct, 1)
				group, err := sprintTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(ct, "Member", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					ct,
					sprintTestRef.cloudTestKit.AuthorizationService,
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
			prepareData: func(sprintTestRef SprintTestRef, requesterUserID uint64) *errs.Error {
				ct := context.Background()
				ct = ctx.NewContextWithUserID(ct, 1)
				group, err := sprintTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(ct, "Member", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					ct,
					sprintTestRef.cloudTestKit.AuthorizationService,
					group.ID,
					teamID,
					authorization.TeamMemberResourceTypeOperations,
					4)
			},
			requesterUserID: 3,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			sprintTestRef, ok := prepareSprintTestRef(t, testCase.toggles)
			if !ok {
				return
			}

			if testCase.prepareData != nil {
				err := testCase.prepareData(sprintTestRef, testCase.requesterUserID)
				if !assert.Nil(t, err) {
					return
				}
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)
			now := time.Now().UTC()
			tx, err := sprintTestRef.transactionFactory.BeginTx(ct, nil)
			if !assert.Nil(t, err) {
				return
			}

			defer tx.Rollback()

			// user := entity.User{
			// 	ID: testCase.requesterUserID,
			// }

			// err = sprintTestRef.userDaoV2.CreateUser(ct, tx, user)
			// if !assert.Nil(t, err) {
			// 	return
			// }

			team := entity.Team{
				ID:            teamID,
				Name:          "test team",
				CreatorUserID: 10,
			}

			err = sprintTestRef.teamDaoV2.CreateTeam(ct, tx, team)
			if !assert.Nil(t, err) {
				return
			}

			sprint := CreateSprintInput{
				StartAt: now,
				EndAt:   now.Add(time.Hour * 24 * 7),
			}

			createdSprint, internalErr := sprintTestRef.sprintService.CreateSprint(ct, teamID, sprint)

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else if !assert.Nil(t, internalErr) {
				return
			}

			assert.Equal(t, createdSprint.OwningTeamID, teamID)
			assert.Equal(t, createdSprint.StartAt, sprint.StartAt)
			assert.Equal(t, createdSprint.EndAt, sprint.EndAt)
		})
	}

}
