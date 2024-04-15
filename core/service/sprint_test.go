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
	"github.com/teamyapp/cloud/libs/network/networktest"
	"github.com/teamyapp/cloud/libs/retry"
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/rpc"
	"github.com/teamyapp/cloud/libs/runtime"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudtx "github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/cloud/testkit"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/dao/daotest"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/instrument/instrumenttest"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/service/servicetest"
	"github.com/teamyapp/teamy-backend/core/transaction"
)

type SprintTestRef struct {
	sprintService         Sprint
	sprintDao             dao.Sprint
	teamDao               dao.Team
	userDao               dao.User
	taskDao               dao.Task
	sprintTaskRelationDao dao.SprintTaskRelation
	sprintParticipantDao  dao.SprintParticipant
	cloudTestKit          testkit.TestKit
	transactionFactory    cloudtx.Factory
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

	internalErr = servicetest.ApplyAuthorizationConfig(cloudTestKit.AuthorizationService)
	require.Nil(t, internalErr)

	testkit.StartServiceInstance(cloudTestKitConfig, virtualNetwork, cloudTestKit.ServiceInstanceRunner)
	apiToken, internalErr := servicetest.GetServiceAccountAPIToken(cloudTestKit.IdentityService)
	require.Nil(t, internalErr)

	prometheus := instrumenttest.NewNoopMetrics()
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
		prometheus,
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
	teamyBackendDB.CreateTable(daotest.TeamTableName)
	teamyBackendDB.CreateTable(daotest.SprintTableName)
	teamyBackendDB.CreateTable(daotest.TeamMemberTableName)
	teamyBackendDB.CreateTable(daotest.SprintTaskRelationTableName)
	teamyBackendDB.CreateTable(daotest.SprintParticipantTableName)
	teamyBackendDB.CreateTable(daotest.TaskTableName)
	teamMemberDao := daotest.NewTeamMember(teamyBackendDB, transactionFactory)
	sprintDao := daotest.NewSprint(teamyBackendDB, transactionFactory)
	taskDao := daotest.NewTask(teamyBackendDB, transactionFactory)
	stateSyncer := realtime.NewStateSyncer(logger, teamMemberDao)
	teamDao := daotest.NewTeam(teamyBackendDB, transactionFactory)
	sprintTaskRelationDao := daotest.NewSprintTaskRelation(teamyBackendDB)
	sprintParticipantDao := daotest.NewSprintParticipant(teamyBackendDB, transactionFactory)
	threadDao := daotest.NewThread(teamyBackendDB)
	userDao := daotest.NewUser(teamyBackendDB, transactionFactory)

	transactionGroupFactory := transaction.NewGroupFactory(logger, prometheus, transactionFactory, stateSyncer)
	sprintService := NewSprint(
		logger,
		transactionGroupFactory,
		cloudClientRegistry,
		stateSyncer,
		authorizer,
		toggles,
		transactionFactory,
		taskDao,
		sprintDao,
		teamDao,
		sprintTaskRelationDao,
		sprintParticipantDao,
		teamMemberDao,
		threadDao,
	)

	return SprintTestRef{
		sprintService:         sprintService,
		cloudTestKit:          cloudTestKit,
		sprintDao:             sprintDao,
		teamDao:               teamDao,
		userDao:               userDao,
		taskDao:               taskDao,
		sprintTaskRelationDao: sprintTaskRelationDao,
		sprintParticipantDao:  sprintParticipantDao,
		transactionFactory:    transactionFactory,
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
					teamID,
					group.ID,
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
					teamID,
					group.ID,
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
					teamID,
					group.ID,
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

			err = sprintTestRef.teamDao.CreateTeam(ct, tx, team)
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
					teamID,
					group.ID,
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
					teamID,
					group.ID,
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
					teamID,
					group.ID,
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
			err = sprintTestRef.teamDao.CreateTeam(ct, tx, team)
			require.Nil(t, err)

			sprint := entity.Sprint{
				ID:           1,
				StartAt:      now,
				EndAt:        now.Add(time.Hour * 24 * 7),
				OwningTeamID: teamID,
			}
			err = sprintTestRef.sprintDao.CreateSprint(ct, tx, sprint)
			require.Nil(t, err)

			participant := entity.SprintParticipant{
				SprintID:        sprint.ID,
				UserID:          testCase.requesterUserID,
				TotalBandwidth:  time.Hour * 24 * 7,
				UnusedBandwidth: time.Hour * 12 * 7,
				CreatedAt:       now,
			}
			err = sprintTestRef.sprintService.sprintParticipantDao.CreateSprintParticipant(ct, tx, participant)
			require.Nil(t, err)

			effort := time.Hour * 12 * 7
			task := entity.Task{
				Goal:          "test task",
				CreatorUserID: 1,
				Status:        entity.TaskStatusTodo,
				IsScheduled:   true,
				IsPlanned:     false,
				OwnerUserID:   &testCase.requesterUserID,
				Effort:        &effort,
			}

			err = sprintTestRef.taskDao.CreateTask(ct, tx, task)
			require.Nil(t, err)

			sprintTaskRelation := entity.SprintTaskRelation{
				SprintID: sprint.ID,
				TaskID:   task.ID,
			}
			err = sprintTestRef.sprintTaskRelationDao.CreateSprintTaskRelation(ct, tx, sprintTaskRelation)
			require.Nil(t, err)

			deletedSprint, internalErr := sprintTestRef.sprintService.DeleteSprint(ct, sprint.ID)
			if testCase.expectedErr != nil {
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			updatedTask, err := sprintTestRef.taskDao.FindTaskByIDWithTx(ct, tx, task.ID)
			require.Nil(t, err)

			_, err = sprintTestRef.sprintParticipantDao.FindParticipantWithTx(ct, tx, participant.SprintID, participant.UserID)
			require.Equal(t, err.Code, errs.NotFound)

			require.Equal(t, deletedSprint.OwningTeamID, teamID)
			require.Equal(t, deletedSprint.StartAt, sprint.StartAt)
			require.Equal(t, deletedSprint.EndAt, sprint.EndAt)
			require.Equal(t, updatedTask.IsScheduled, false)
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
					teamID,
					group.ID,
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
					teamID,
					group.ID,
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
					teamID,
					group.ID,
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

			err = sprintTestRef.sprintDao.CreateSprint(ct, tx, sprint1)
			require.Nil(t, err)

			err = sprintTestRef.sprintDao.CreateSprint(ct, tx, sprint2)
			require.Nil(t, err)

			team := entity.Team{
				ID:            teamID,
				Name:          "test team",
				CreatorUserID: 10,
				CreatedAt:     now,
			}

			err = sprintTestRef.teamDao.CreateTeam(ct, tx, team)
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

func TestSprintService_GetTeamActiveSprint(t *testing.T) {
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
				ct = ctx.NewContextWithUserID(ct, servicetest.AutomationUserID)
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
					teamID,
					group.ID,
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
				ct = ctx.NewContextWithUserID(ct, servicetest.AutomationUserID)
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
					teamID,
					group.ID,
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
				ct = ctx.NewContextWithUserID(ct, servicetest.AutomationUserID)
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
			prepareData: func(sprintTestRef SprintTestRef, requesterUserID uint64) *errs.Error {
				ct := context.Background()
				ct = ctx.NewContextWithUserID(ct, servicetest.AutomationUserID)
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
					teamID,
					group.ID,
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

			err := servicetest.AddAllTeamPermissions(
				sprintTestRef.cloudTestKit.AuthorizationService,
				teamID,
				servicetest.AutomationUserID)
			require.Nil(t, err)

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

			sprint1, err := sprintTestRef.sprintService.CreateSprint(
				servicetest.AutomationCtx,
				teamID,
				CreateSprintInput{
					StartAt: now,
					EndAt:   now.Add(time.Hour * 24 * 7),
				})
			require.Nil(t, err)

			team := entity.Team{
				ID:             teamID,
				ActiveSprintID: &sprint1.ID,
				Name:           "test team",
				CreatorUserID:  10,
				CreatedAt:      now,
			}

			err = sprintTestRef.teamDao.CreateTeam(ct, tx, team)
			require.Nil(t, err)

			activeSprint, internalErr := sprintTestRef.sprintService.GetActiveSprint(ct, teamID)
			if testCase.expectedErr != nil {
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			require.Equal(t, sprint1.ID, activeSprint.ID)
		})
	}

}
