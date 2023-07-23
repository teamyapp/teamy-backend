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
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/daov2/daotestv2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/service/servicetest"
)

type SprintTestRef struct {
	sprintService           Sprint
	sprintDaoV2             daov2.Sprint
	teamDaoV2               daov2.Team
	userDaoV2               daov2.User
	taskDaoV2               daov2.Task
	sprintTaskRelationDaoV2 daov2.SprintTaskRelation
	sprintParticipantDaoV2  daov2.SprintParticipant
	cloudTestKit            testkit.TestKit
	transactionFactory      transaction.Factory
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
	require.Nil(t, internalErr)

	testkit.StartServiceInstance(cloudTestKitConfig, virtualNetwork, cloudTestKit.ServiceInstanceRunner)
	apiToken, internalErr := servicetest.GetServiceAccountAPIToken(cloudTestKit.IdentityService)
	require.Nil(t, internalErr)

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
	require.Nil(t, err)

	authorizer := client.NewAuthorizer(logger, cloudClientRegistry)
	transactionFactory := transaction.NewFactory(nil)
	teamyBackendDB := dbtest.NewInMemoryDB()
	teamyBackendDB.CreateTable(daotestv2.TeamTableName)
	teamyBackendDB.CreateTable(daotestv2.SprintTableName)
	teamyBackendDB.CreateTable(daotestv2.TeamMemberTableName)
	teamyBackendDB.CreateTable(daotestv2.SprintTaskRelationTableName)
	teamyBackendDB.CreateTable(daotestv2.SprintParticipantTableName)
	teamyBackendDB.CreateTable(daotestv2.TaskTableName)
	teamMemberDaoV2 := daotestv2.NewTeamMember(teamyBackendDB, transactionFactory)
	sprintDaoV2 := daotestv2.NewSprint(teamyBackendDB, transactionFactory)
	taskDaoV2 := daotestv2.NewTask(teamyBackendDB, transactionFactory)
	stateSyncer := realtime.NewStateSyncer(logger, teamMemberDaoV2)
	teamDaoV2 := daotestv2.NewTeam(teamyBackendDB, transactionFactory)
	sprintTaskRelationDaoV2 := daotestv2.NewSprintTaskRelation(teamyBackendDB)
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
		taskDaoV2,
		sprintDaoV2,
		teamDaoV2,
		sprintTaskRelationDaoV2,
		sprintParticipantDaoV2,
		teamMemberDaoV2,
		threadDaoV2,
	)

	return SprintTestRef{
		sprintService:           sprintService,
		cloudTestKit:            cloudTestKit,
		sprintDaoV2:             sprintDaoV2,
		teamDaoV2:               teamDaoV2,
		userDaoV2:               userDaoV2,
		taskDaoV2:               taskDaoV2,
		sprintTaskRelationDaoV2: sprintTaskRelationDaoV2,
		sprintParticipantDaoV2:  sprintParticipantDaoV2,
		transactionFactory:      transactionFactory,
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
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)
			now := time.Now().UTC()
			tx, err := sprintTestRef.transactionFactory.BeginTx(ct, nil)
			require.Nil(t, err)

			defer tx.Rollback()

			team := entity.Team{
				ID:            teamID,
				Name:          "test team",
				CreatorUserID: 10,
			}

			err = sprintTestRef.teamDaoV2.CreateTeam(ct, tx, team)
			require.Nil(t, err)

			sprint := CreateSprintInput{
				StartAt: now,
				EndAt:   now.Add(time.Hour * 24 * 7),
			}

			createdSprint, internalErr := sprintTestRef.sprintService.CreateSprint(ct, teamID, sprint)

			if testCase.expectedErr != nil {
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			require.Equal(t, createdSprint.OwningTeamID, teamID)
			require.Equal(t, createdSprint.StartAt, sprint.StartAt)
			require.Equal(t, createdSprint.EndAt, sprint.EndAt)
		})
	}

}

func TestSprintService_DeleteSprint(t *testing.T) {
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
			// TODO: turn it back on when ApplyAuthorizationConfig is implemented.
			toggles: feature.Toggles{
				EnableAuthorization: false,
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
			// TODO: turn it back on when ApplyAuthorizationConfig is implemented.
			toggles: feature.Toggles{
				EnableAuthorization: false,
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
			// TODO: turn it back on when ApplyAuthorizationConfig is implemented.
			toggles: feature.Toggles{
				EnableAuthorization: false,
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
		// TODO: Pending on authorization config api to be ready
		// {
		// 	name: "permission denied when user is not in team",
		// 	toggles: feature.Toggles{
		// 		EnableAuthorization: false,
		// 	},
		// 	prepareData: func(sprintTestRef SprintTestRef, requesterUserID uint64) *errs.Error {
		// 		ct := context.Background()
		// 		ct = ctx.NewContextWithUserID(ct, 1)
		// 		group, err := sprintTestRef.
		// 			cloudTestKit.
		// 			AuthorizationService.
		// 			CreateUserGroup(ct, "Member", nil)
		// 		if err != nil {
		// 			return err
		// 		}

		// 		return servicetest.AddTeamPermission(
		// 			ct,
		// 			sprintTestRef.cloudTestKit.AuthorizationService,
		// 			group.ID,
		// 			teamID,
		// 			authorization.TeamMemberResourceTypeOperations,
		// 			4)
		// 	},
		// 	requesterUserID: 3,
		// 	expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		// },
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
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)
			now := time.Now().UTC()
			tx, err := sprintTestRef.transactionFactory.BeginTx(ct, nil)
			require.Nil(t, err)

			defer tx.Rollback()
			team := entity.Team{
				ID:            teamID,
				Name:          "test team",
				CreatorUserID: 10,
			}
			err = sprintTestRef.teamDaoV2.CreateTeam(ct, tx, team)
			require.Nil(t, err)

			sprint := entity.Sprint{
				ID:           1,
				StartAt:      now,
				EndAt:        now.Add(time.Hour * 24 * 7),
				OwningTeamID: teamID,
			}
			err = sprintTestRef.sprintDaoV2.CreateSprint(ct, tx, sprint)
			require.Nil(t, err)

			participant := entity.SprintParticipant{
				SprintID:        sprint.ID,
				UserID:          testCase.requesterUserID,
				TotalBandwidth:  time.Hour * 24 * 7,
				UnusedBandwidth: time.Hour * 12 * 7,
				CreatedAt:       now,
			}
			err = sprintTestRef.sprintService.sprintParticipantDaoV2.CreateSprintParticipant(ct, tx, participant)
			require.Nil(t, err)

			effort := time.Hour * 12 * 7
			task := entity.Task{
				Goal:          "test task",
				CreatorUserID: 1,
				Status:        entity.TaskStatusTodo,
				IsPlanned:     true,
				OwnerUserID:   &testCase.requesterUserID,
				Effort:        &effort,
			}

			err = sprintTestRef.taskDaoV2.CreateTask(ct, tx, task)
			require.Nil(t, err)

			sprintTaskRelation := entity.SprintTaskRelation{
				SprintID: sprint.ID,
				TaskID:   task.ID,
			}
			err = sprintTestRef.sprintTaskRelationDaoV2.CreateSprintTaskRelation(ct, tx, sprintTaskRelation)
			require.Nil(t, err)

			deletedSprint, internalErr := sprintTestRef.sprintService.DeleteSprint(ct, sprint.ID)
			if testCase.expectedErr != nil {
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			updatedTask, err := sprintTestRef.taskDaoV2.FindTaskByIDWithTx(ct, tx, task.ID)
			require.Nil(t, err)

			_, err = sprintTestRef.sprintParticipantDaoV2.FindParticipantWithTx(ct, tx, participant.SprintID, participant.UserID)
			require.Equal(t, err.Code, errs.NotFound)

			require.Equal(t, deletedSprint.OwningTeamID, teamID)
			require.Equal(t, deletedSprint.StartAt, sprint.StartAt)
			require.Equal(t, deletedSprint.EndAt, sprint.EndAt)
			require.Equal(t, updatedTask.IsPlanned, false)
			require.Equal(t, updatedTask.Status, entity.TaskStatusTodo)
		})
	}
}

func TestSprintService_SetTeamActiveSprint(t *testing.T) {
	var teamID uint64 = 1
	var sprintID1 uint64 = 2
	var sprintID2 uint64 = 3
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
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)
			now := time.Now().UTC()
			tx, err := sprintTestRef.transactionFactory.BeginTx(ct, nil)
			require.Nil(t, err)

			defer tx.Rollback()

			sprint1 := entity.Sprint{
				ID:           sprintID1,
				StartAt:      now,
				EndAt:        now.Add(time.Hour * 24 * 7),
				CreatedAt:    now,
				OwningTeamID: teamID,
			}

			sprint2 := entity.Sprint{
				ID:           sprintID2,
				StartAt:      now.Add(time.Hour * 24),
				EndAt:        now.Add(time.Hour * 24 * 8),
				CreatedAt:    now,
				OwningTeamID: teamID,
			}

			err = sprintTestRef.sprintDaoV2.CreateSprint(ct, tx, sprint1)
			require.Nil(t, err)

			err = sprintTestRef.sprintDaoV2.CreateSprint(ct, tx, sprint2)
			require.Nil(t, err)

			team := entity.Team{
				ID:            teamID,
				Name:          "test team",
				CreatorUserID: 10,
				CreatedAt:     now,
			}

			err = sprintTestRef.teamDaoV2.CreateTeam(ct, tx, team)
			require.Nil(t, err)

			updatedTeam, internalErr := sprintTestRef.sprintService.SetTeamActiveSprint(ct, teamID, sprintID1)
			if testCase.expectedErr != nil {
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			require.Equal(t, sprintID1, *updatedTeam.ActiveSprintID)

			updatedTeam, internalErr = sprintTestRef.sprintService.SetTeamActiveSprint(ct, teamID, sprintID2)
			if testCase.expectedErr != nil {
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			}

			require.Nil(t, internalErr)
			require.Equal(t, sprintID2, *updatedTeam.ActiveSprintID)
		})
	}

}

func TestSprintServiceV2_GetTeamActiveSprint(t *testing.T) {
	var teamID uint64 = 1
	var sprintID1 uint64 = 2
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
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)
			now := time.Now().UTC()
			tx, err := sprintTestRef.transactionFactory.BeginTx(ct, nil)
			require.Nil(t, err)

			defer tx.Rollback()

			sprint1 := entity.Sprint{
				ID:           sprintID1,
				StartAt:      now,
				EndAt:        now.Add(time.Hour * 24 * 7),
				CreatedAt:    now,
				OwningTeamID: teamID,
			}

			err = sprintTestRef.sprintDaoV2.CreateSprint(ct, tx, sprint1)
			require.Nil(t, err)

			team := entity.Team{
				ID:             teamID,
				ActiveSprintID: &sprintID1,
				Name:           "test team",
				CreatorUserID:  10,
				CreatedAt:      now,
			}

			err = sprintTestRef.teamDaoV2.CreateTeam(ct, tx, team)
			require.Nil(t, err)

			activeSprint, internalErr := sprintTestRef.sprintService.GetActiveSprint(ct, teamID)
			if testCase.expectedErr != nil {
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			require.Equal(t, sprintID1, activeSprint.ID)
		})
	}

}
