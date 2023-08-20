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
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/dao/daotest"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/service/servicetest"
)

type InvitationTestRef struct {
	invitationService  Invitation
	invitationDao      dao.Invitation
	transactionFactory transaction.Factory
	cloudTestKit       testkit.TestKit
}

func prepareInvitationTestRef(t *testing.T, toggles feature.Toggles) InvitationTestRef {
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
	teamyBackendDB.CreateTable(daotest.InvitationTableName)
	teamyBackendDB.CreateTable(daotest.TeamMemberTableName)
	teamyBackendDB.CreateTable(daotest.SprintTableName)

	teamMemberDao := daotest.NewTeamMember(teamyBackendDB, transactionFactory)
	stateSyncer := realtime.NewStateSyncer(logger, teamMemberDao)

	invitationDao := daotest.NewInvitation(teamyBackendDB, transactionFactory)
	sprintParticipantDao := daotest.NewSprintParticipant(teamyBackendDB, transactionFactory)
	sprintDao := daotest.NewSprint(teamyBackendDB, transactionFactory)
	invitationService := NewInvitation(
		logger,
		cloudClientRegistry,
		authorizer,
		toggles,
		stateSyncer,
		transactionFactory,
		invitationDao,
		teamMemberDao,
		sprintParticipantDao,
		sprintDao,
	)
	return InvitationTestRef{
		invitationService:  invitationService,
		invitationDao:      invitationDao,
		transactionFactory: transactionFactory,
		cloudTestKit:       cloudTestKit,
	}
}

func TestInvitationService_FindInvitationsInTeam(t *testing.T) {
	var teamID uint64 = 1
	testCases := []struct {
		name               string
		toggles            feature.Toggles
		prepareData        func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error
		requesterUserID    uint64
		maxInvitationCount int
		expectedErr        *errs.Error
	}{
		{
			name: "succeed when user is team owner",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				ct := context.Background()
				ct = ctx.NewContextWithUserID(ct, servicetest.AutomationUserID)
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(ct, "Owner", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					ct,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					teamID,
					group.ID,
					authorization.TeamOwnerResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID:    1,
			maxInvitationCount: 3,
			expectedErr:        nil,
		},
		{
			name: "succeed when user is team admin",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(servicetest.AutomationCtx, "Admin", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					servicetest.AutomationCtx,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					teamID,
					group.ID,
					authorization.TeamAdminResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID:    1,
			maxInvitationCount: 3,
			expectedErr:        nil,
		},
		{
			name: "hide invitations when user is team member",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(servicetest.AutomationCtx, "Member", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					servicetest.AutomationCtx,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					teamID,
					group.ID,
					authorization.TeamMemberResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID:    1,
			maxInvitationCount: 0,
			expectedErr:        nil,
		},
		{
			name: "permission denied when user is not in team",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(servicetest.AutomationCtx, "Member", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					servicetest.AutomationCtx,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					teamID,
					group.ID,
					authorization.TeamMemberResourceTypeOperations,
					4)
			},
			requesterUserID:    3,
			maxInvitationCount: 0,
			expectedErr:        errs.NewError(errs.PermissionDenied, "permission denied"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			invitationRef := prepareInvitationTestRef(t, testCase.toggles)
			if testCase.prepareData != nil {
				err := testCase.prepareData(invitationRef, testCase.requesterUserID)
				require.Nil(t, err)
			}

			err := servicetest.AddAllTeamPermissions(
				invitationRef.cloudTestKit.AuthorizationService,
				teamID,
				servicetest.AutomationUserID)
			require.Nil(t, err)

			var receiverFirstName = "Test_FirstName"
			var receiverLastName = "Test_LastName"
			var receiverEmail = "test@teamyapp.com"
			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			now := time.Now().UTC()
			invitation1, err := invitationRef.invitationService.CreateInvitation(
				servicetest.AutomationCtx,
				teamID,
				CreateInvitationInput{
					ReceiverFirstName: &receiverFirstName,
					ReceiverLastName:  &receiverLastName,
					ReceiverEmail:     &receiverEmail,
					ExpireAt:          now.Add(2 * time.Hour),
				})
			require.Nil(t, err)

			_, err = invitationRef.invitationService.CreateInvitation(
				servicetest.AutomationCtx,
				teamID,
				CreateInvitationInput{
					ReceiverFirstName: &receiverFirstName,
					ReceiverLastName:  &receiverLastName,
					ReceiverEmail:     &receiverEmail,
					ExpireAt:          now.Add(2 * time.Hour),
				})
			require.Nil(t, err)

			_, err = invitationRef.invitationService.CreateInvitation(
				servicetest.AutomationCtx,
				teamID,
				CreateInvitationInput{
					ReceiverFirstName: &receiverFirstName,
					ReceiverLastName:  &receiverLastName,
					ReceiverEmail:     &receiverEmail,
					ExpireAt:          now.Add(2 * time.Hour),
				})
			require.Nil(t, err)

			filter1 := InvitationFilter{}
			invitationsFound, internalErr := invitationRef.invitationService.FindInvitationsInTeam(ct, teamID, &filter1)
			if testCase.expectedErr != nil {
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			require.Equal(t, testCase.maxInvitationCount, len(invitationsFound))

			filter2 := InvitationFilter{InvitationID: &invitation1.ID}
			invitationsFound, internalErr = invitationRef.invitationService.FindInvitationsInTeam(ct, teamID, &filter2)
			if testCase.expectedErr != nil {
				require.Equal(t, testCase.expectedErr.Code, internalErr.Code)
				return
			} else {
				require.Nil(t, internalErr)
			}

			if testCase.maxInvitationCount > 0 {
				require.Equal(t, 1, len(invitationsFound))
				require.Equal(t, invitation1.ID, invitationsFound[0].ID)
			}
		})
	}
}

func TestInvitationService_FindInvitations(t *testing.T) {
	var team1ID uint64 = 1
	var team2ID uint64 = 2
	testCases := []struct {
		name               string
		toggles            feature.Toggles
		prepareData        func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error
		requesterUserID    uint64
		maxInvitationCount int
		expectedErr        *errs.Error
	}{
		{
			name: "succeed when user is team 1 owner",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				ct := context.Background()
				ct = ctx.NewContextWithUserID(ct, servicetest.AutomationUserID)
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(ct, "Owner", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					ct,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					team1ID,
					group.ID,
					authorization.TeamOwnerResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID:    1,
			maxInvitationCount: 3,
			expectedErr:        nil,
		},
		{
			name: "succeed when user is team 1 admin",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(servicetest.AutomationCtx, "Admin", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					servicetest.AutomationCtx,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					team1ID,
					group.ID,
					authorization.TeamAdminResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID:    1,
			maxInvitationCount: 3,
			expectedErr:        nil,
		},
		{
			name: "hide invitations when user is team 1 member",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(servicetest.AutomationCtx, "Member", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					servicetest.AutomationCtx,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					team1ID,
					group.ID,
					authorization.TeamMemberResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID:    1,
			maxInvitationCount: 0,
			expectedErr:        nil,
		},
		{
			name: "hide invitations when user is not in team",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(servicetest.AutomationCtx, "Member", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					servicetest.AutomationCtx,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					team1ID,
					group.ID,
					authorization.TeamMemberResourceTypeOperations,
					4)
			},
			requesterUserID:    3,
			maxInvitationCount: 0,
			expectedErr:        nil,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			invitationRef := prepareInvitationTestRef(t, testCase.toggles)
			if testCase.prepareData != nil {
				err := testCase.prepareData(invitationRef, testCase.requesterUserID)
				require.Nil(t, err)
			}

			err := servicetest.AddAllTeamPermissions(
				invitationRef.cloudTestKit.AuthorizationService,
				team1ID,
				servicetest.AutomationUserID)
			require.Nil(t, err)

			err = servicetest.AddAllTeamPermissions(
				invitationRef.cloudTestKit.AuthorizationService,
				team2ID,
				servicetest.AutomationUserID)
			require.Nil(t, err)

			now := time.Now().UTC()
			var receiverFirstName = "Test_FirstName"
			var receiverLastName = "Test_LastName"
			var receiverEmail = "test@test.com"
			invitation1, err := invitationRef.invitationService.CreateInvitation(
				servicetest.AutomationCtx,
				team1ID,
				CreateInvitationInput{
					ReceiverFirstName: &receiverFirstName,
					ReceiverLastName:  &receiverLastName,
					ReceiverEmail:     &receiverEmail,
					ExpireAt:          now.Add(2 * time.Hour),
				})
			require.Nil(t, err)

			_, err = invitationRef.invitationService.CreateInvitation(
				servicetest.AutomationCtx,
				team1ID,
				CreateInvitationInput{
					ReceiverFirstName: &receiverFirstName,
					ReceiverLastName:  &receiverLastName,
					ReceiverEmail:     &receiverEmail,
					ExpireAt:          now.Add(2 * time.Hour),
				})
			require.Nil(t, err)

			_, err = invitationRef.invitationService.CreateInvitation(
				servicetest.AutomationCtx,
				team1ID,
				CreateInvitationInput{
					ReceiverFirstName: &receiverFirstName,
					ReceiverLastName:  &receiverLastName,
					ReceiverEmail:     &receiverEmail,
					ExpireAt:          now.Add(2 * time.Hour),
				})
			require.Nil(t, err)

			_, err = invitationRef.invitationService.CreateInvitation(
				servicetest.AutomationCtx,
				team2ID,
				CreateInvitationInput{
					ReceiverFirstName: &receiverFirstName,
					ReceiverLastName:  &receiverLastName,
					ReceiverEmail:     &receiverEmail,
					ExpireAt:          now.Add(2 * time.Hour),
				})
			require.Nil(t, err)

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			filter1 := InvitationFilter{}
			invitationsFound, internalErr := invitationRef.invitationService.FindInvitations(ct, &filter1)
			require.Nil(t, internalErr)
			require.Equal(t, testCase.maxInvitationCount, len(invitationsFound))

			filter2 := InvitationFilter{InvitationID: &invitation1.ID}
			invitationsFound, internalErr = invitationRef.invitationService.FindInvitations(ct, &filter2)
			require.Nil(t, internalErr)

			if testCase.maxInvitationCount > 0 {
				require.Equal(t, 1, len(invitationsFound))
				require.Equal(t, invitation1.ID, invitationsFound[0].ID)
			}
		})
	}
}

func TestInvitationService_CreateInvitation(t *testing.T) {
	var team1ID uint64 = 1
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed when user is team 1 owner",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				ct := context.Background()
				ct = ctx.NewContextWithUserID(ct, servicetest.AutomationUserID)
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(ct, "Owner", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					ct,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					team1ID,
					group.ID,
					authorization.TeamOwnerResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID: 1,
			expectedErr:     nil,
		},
		{
			name: "succeed when user is team 1 admin",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(servicetest.AutomationCtx, "Admin", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					servicetest.AutomationCtx,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					team1ID,
					group.ID,
					authorization.TeamAdminResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID: 1,
			expectedErr:     nil,
		},
		{
			name: "permission denied when user is team 1 member",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(servicetest.AutomationCtx, "Member", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					servicetest.AutomationCtx,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					team1ID,
					group.ID,
					authorization.TeamMemberResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID: 1,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
		{
			name: "permission denied when user is not team 1 member",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(servicetest.AutomationCtx, "Member", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					servicetest.AutomationCtx,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					team1ID,
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
			invitationRef := prepareInvitationTestRef(t, testCase.toggles)
			if testCase.prepareData != nil {
				err := testCase.prepareData(invitationRef, testCase.requesterUserID)
				require.Nil(t, err)
			}

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			now := time.Now().UTC()
			var receiverFirstName = "Test_FirstName"
			var receiverLastName = "Test_LastName"
			var receiverEmail = "test@test.com"
			_, err := invitationRef.invitationService.CreateInvitation(
				ct,
				team1ID,
				CreateInvitationInput{
					ReceiverFirstName: &receiverFirstName,
					ReceiverLastName:  &receiverLastName,
					ReceiverEmail:     &receiverEmail,
					ExpireAt:          now.Add(2 * time.Hour),
				})
			if testCase.expectedErr != nil {
				require.Equal(t, testCase.expectedErr.Code, err.Code)
				return
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestInvitationService_UpdateInvitation(t *testing.T) {
	var team1ID uint64 = 1
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed when user is team 1 owner",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				ct := context.Background()
				ct = ctx.NewContextWithUserID(ct, servicetest.AutomationUserID)
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(ct, "Owner", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					ct,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					team1ID,
					group.ID,
					authorization.TeamOwnerResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID: 1,
			expectedErr:     nil,
		},
		{
			name: "succeed when user is team 1 admin",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(servicetest.AutomationCtx, "Admin", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					servicetest.AutomationCtx,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					team1ID,
					group.ID,
					authorization.TeamAdminResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID: 1,
			expectedErr:     nil,
		},
		{
			name: "permission denied when user is team 1 member",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(servicetest.AutomationCtx, "Member", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					servicetest.AutomationCtx,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					team1ID,
					group.ID,
					authorization.TeamMemberResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID: 1,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
		{
			name: "permission denied when user is not team 1 member",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(servicetest.AutomationCtx, "Member", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					servicetest.AutomationCtx,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					team1ID,
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
			invitationRef := prepareInvitationTestRef(t, testCase.toggles)
			if testCase.prepareData != nil {
				err := testCase.prepareData(invitationRef, testCase.requesterUserID)
				require.Nil(t, err)
			}

			err := servicetest.AddAllTeamPermissions(
				invitationRef.cloudTestKit.AuthorizationService,
				team1ID,
				servicetest.AutomationUserID)
			require.Nil(t, err)

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)

			now := time.Now().UTC()
			var receiverFirstName = "Test_FirstName"
			var receiverLastName = "Test_LastName"
			var receiverEmail = "test@test.com"
			invitation, err := invitationRef.invitationService.CreateInvitation(
				servicetest.AutomationCtx,
				team1ID,
				CreateInvitationInput{
					ReceiverFirstName: &receiverFirstName,
					ReceiverLastName:  &receiverLastName,
					ReceiverEmail:     &receiverEmail,
					ExpireAt:          now.Add(2 * time.Hour),
				})
			require.Nil(t, err)

			expiredAt := now.Add(3 * time.Hour)
			updatedFirstName := "Updated_FirstName"
			updatedLastName := "Updated_LastName"
			input := UpdateInvitationInput{
				ExpireAt:          expiredAt,
				ReceiverFirstName: &updatedFirstName,
				ReceiverLastName:  &updatedLastName,
			}
			updated, err := invitationRef.invitationService.UpdateInvitation(ct, invitation.ID, input)
			if testCase.expectedErr != nil {
				require.Equal(t, testCase.expectedErr.Code, err.Code)
				return
			} else {
				require.Nil(t, err)
			}

			require.Equal(t, invitation.ID, updated.ID)
			require.Equal(t, invitation.SenderUserID, updated.SenderUserID)
			require.Equal(t, invitation.TeamID, updated.TeamID)
			require.Equal(t, invitation.CreatedAt, updated.CreatedAt)
			require.Equal(t, invitation.Status, updated.Status)
			require.Equal(t, input.ReceiverFirstName, updated.ReceiverFirstName)
			require.Equal(t, input.ReceiverLastName, updated.ReceiverLastName)
			require.Equal(t, input.ExpireAt, updated.ExpireAt)
		})
	}
}

func TestInvitationService_DeleteInvitation(t *testing.T) {
	var team1ID uint64 = 1
	testCases := []struct {
		name            string
		toggles         feature.Toggles
		prepareData     func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error
		requesterUserID uint64
		expectedErr     *errs.Error
	}{
		{
			name: "succeed when user is team 1 owner",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				ct := context.Background()
				ct = ctx.NewContextWithUserID(ct, servicetest.AutomationUserID)
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(ct, "Owner", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					ct,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					team1ID,
					group.ID,
					authorization.TeamOwnerResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID: 1,
			expectedErr:     nil,
		},
		{
			name: "succeed when user is team 1 admin",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(servicetest.AutomationCtx, "Admin", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					servicetest.AutomationCtx,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					team1ID,
					group.ID,
					authorization.TeamAdminResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID: 1,
			expectedErr:     nil,
		},
		{
			name: "permission denied when user is team 1 member",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(servicetest.AutomationCtx, "Member", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					servicetest.AutomationCtx,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					team1ID,
					group.ID,
					authorization.TeamMemberResourceTypeOperations,
					requesterUserID)
			},
			requesterUserID: 1,
			expectedErr:     errs.NewError(errs.PermissionDenied, "permission denied"),
		},
		{
			name: "permission denied when user is not team 1 member",
			toggles: feature.Toggles{
				EnableAuthorization: true,
			},
			prepareData: func(invitationTestRef InvitationTestRef, requesterUserID uint64) *errs.Error {
				group, err := invitationTestRef.
					cloudTestKit.
					AuthorizationService.
					CreateUserGroup(servicetest.AutomationCtx, "Member", nil)
				if err != nil {
					return err
				}

				return servicetest.AddTeamPermission(
					servicetest.AutomationCtx,
					invitationTestRef.
						cloudTestKit.
						AuthorizationService,
					team1ID,
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
			invitationRef := prepareInvitationTestRef(t, testCase.toggles)
			if testCase.prepareData != nil {
				err := testCase.prepareData(invitationRef, testCase.requesterUserID)
				require.Nil(t, err)
			}

			err := servicetest.AddAllTeamPermissions(
				invitationRef.cloudTestKit.AuthorizationService,
				team1ID,
				servicetest.AutomationUserID)
			require.Nil(t, err)

			now := time.Now().UTC()
			var receiverFirstName = "Test_FirstName"
			var receiverLastName = "Test_LastName"
			var receiverEmail = "test@test.com"
			invitation, err := invitationRef.invitationService.CreateInvitation(
				servicetest.AutomationCtx,
				team1ID,
				CreateInvitationInput{
					ReceiverFirstName: &receiverFirstName,
					ReceiverLastName:  &receiverLastName,
					ReceiverEmail:     &receiverEmail,
					ExpireAt:          now.Add(2 * time.Hour),
				})
			require.Nil(t, err)

			ct := context.Background()
			ct = ctx.NewContextWithUserID(ct, testCase.requesterUserID)
			deleted, err := invitationRef.invitationService.DeleteInvitation(ct, invitation.ID)
			if testCase.expectedErr != nil {
				require.Equal(t, testCase.expectedErr.Code, err.Code)
				return
			} else {
				require.Nil(t, err)
			}

			require.Equal(t, invitation.ID, deleted.ID)
			require.Equal(t, invitation.SenderUserID, deleted.SenderUserID)
			require.Equal(t, invitation.TeamID, deleted.TeamID)
			require.Equal(t, invitation.CreatedAt, deleted.CreatedAt)
			require.Equal(t, invitation.Status, deleted.Status)
			require.Equal(t, invitation.ReceiverFirstName, deleted.ReceiverFirstName)
			require.Equal(t, invitation.ReceiverLastName, deleted.ReceiverLastName)
			require.Equal(t, invitation.ExpireAt, deleted.ExpireAt)
		})
	}
}

func TestInvitationService_AcceptInvitation(t *testing.T) {
	var team1ID uint64 = 1
	var requesterUserID uint64 = 1
	invitationRef := prepareInvitationTestRef(t, feature.Toggles{
		EnableAuthorization: true,
	})
	err := servicetest.AddAllTeamPermissions(
		invitationRef.cloudTestKit.AuthorizationService,
		team1ID,
		servicetest.AutomationUserID)
	require.Nil(t, err)

	now := time.Now().UTC()
	var receiverFirstName = "Test_FirstName"
	var receiverLastName = "Test_LastName"
	var receiverEmail = "test@test.com"
	invitation, err := invitationRef.invitationService.CreateInvitation(
		servicetest.AutomationCtx,
		team1ID,
		CreateInvitationInput{
			ReceiverFirstName: &receiverFirstName,
			ReceiverLastName:  &receiverLastName,
			ReceiverEmail:     &receiverEmail,
			ExpireAt:          now.Add(2 * time.Hour),
		})
	require.Nil(t, err)

	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)
	accepted, err := invitationRef.invitationService.AcceptInvitation(ct, invitation.ID, "random")
	require.NotNil(t, err)
	require.Equal(t, errs.InvalidArgument, err.Code)

	accepted, err = invitationRef.invitationService.AcceptInvitation(ct, invitation.ID, invitation.Code)
	require.Nil(t, err)
	require.Equal(t, entity.InvitationStatusAccepted, accepted.Status)
}

func TestInvitationService_DeclineInvitation(t *testing.T) {
	var team1ID uint64 = 1
	var requesterUserID uint64 = 1
	invitationRef := prepareInvitationTestRef(t, feature.Toggles{
		EnableAuthorization: true,
	})
	err := servicetest.AddAllTeamPermissions(
		invitationRef.cloudTestKit.AuthorizationService,
		team1ID,
		servicetest.AutomationUserID)
	require.Nil(t, err)

	now := time.Now().UTC()
	var receiverFirstName = "Test_FirstName"
	var receiverLastName = "Test_LastName"
	var receiverEmail = "test@test.com"
	invitation, err := invitationRef.invitationService.CreateInvitation(
		servicetest.AutomationCtx,
		team1ID,
		CreateInvitationInput{
			ReceiverFirstName: &receiverFirstName,
			ReceiverLastName:  &receiverLastName,
			ReceiverEmail:     &receiverEmail,
			ExpireAt:          now.Add(2 * time.Hour),
		})
	require.Nil(t, err)

	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)
	accepted, err := invitationRef.invitationService.DeclineInvitation(ct, invitation.ID, "random")
	require.NotNil(t, err)
	require.Equal(t, errs.InvalidArgument, err.Code)

	accepted, err = invitationRef.invitationService.DeclineInvitation(ct, invitation.ID, invitation.Code)
	require.Nil(t, err)
	require.Equal(t, entity.InvitationStatusDeclined, accepted.Status)
}
