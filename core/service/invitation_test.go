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
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/dao/daotest"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type InvitationTestRef struct {
	invitationService  Invitation
	invitationDao      dao.Invitation
	transactionFactory transaction.Factory
}

func prepareInvitationTestRef(t *testing.T, toggles feature.Toggles) (InvitationTestRef, bool) {
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
	}, true
}

func TestInvitationService_FindInvitationsInTeam(t *testing.T) {
	invitationRef, ok := prepareInvitationTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var invitationID1 uint64 = 12
	var invitationID2 uint64 = 13
	var invitationID3 uint64 = 14
	var receiverFirstName = "Test_FirstName"
	var receiverLastName = "Test_LastName"
	var receiverEmail = "test@test.com"
	var teamID1 uint64 = 5
	var teamID2 uint64 = 6
	var code1 = "Test code 1"
	var code2 = "Test code 2"
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := invitationRef.transactionFactory.BeginTx(ct, nil)
	require.Nil(t, err)

	defer tx.Rollback()

	now := time.Now().UTC()
	invitation1 := entity.Invitation{
		ID:                invitationID1,
		SenderUserID:      requesterUserID,
		ReceiverFirstName: &receiverFirstName,
		ReceiverLastName:  &receiverLastName,
		ReceiverEmail:     &receiverEmail,
		TeamID:            teamID1,
		ExpireAt:          now.Add(2 * time.Hour),
		Status:            entity.InvitationStatusPending,
		Code:              code2,
		CreatedAt:         now,
		UpdatedAt:         &now,
	}

	invitation2 := entity.Invitation{
		ID:                invitationID2,
		SenderUserID:      requesterUserID,
		ReceiverFirstName: &receiverFirstName,
		ReceiverLastName:  &receiverLastName,
		ReceiverEmail:     &receiverEmail,
		TeamID:            teamID2,
		ExpireAt:          now.Add(2 * time.Hour),
		Status:            entity.InvitationStatusPending,
		Code:              code1,
		CreatedAt:         now,
		UpdatedAt:         &now,
	}

	invitation3 := entity.Invitation{
		ID:                invitationID3,
		SenderUserID:      requesterUserID,
		ReceiverFirstName: &receiverFirstName,
		ReceiverLastName:  &receiverLastName,
		ReceiverEmail:     &receiverEmail,
		TeamID:            teamID2,
		ExpireAt:          now.Add(2 * time.Hour),
		Status:            entity.InvitationStatusPending,
		Code:              code2,
		CreatedAt:         now,
		UpdatedAt:         &now,
	}

	require.Nil(t, invitationRef.invitationDao.CreateInvitation(ct, tx, invitation1))
	require.Nil(t, invitationRef.invitationDao.CreateInvitation(ct, tx, invitation2))
	require.Nil(t, invitationRef.invitationDao.CreateInvitation(ct, tx, invitation3))

	filter1 := InvitationFilter{InvitationID: &invitationID2, Code: &code1}
	invitationsFound, internalErr := invitationRef.invitationService.FindInvitationsInTeam(ct, teamID2, &filter1)
	require.Nil(t, internalErr)
	require.Equal(t, 1, len(invitationsFound))
	require.Equal(t, invitation2.ID, invitationsFound[0].ID)
	require.True(t, areInvitationsEqual(invitation2, invitationsFound[0]))

	filter2 := InvitationFilter{InvitationID: &invitationID2, Code: &code2}
	invitationsFound, internalErr = invitationRef.invitationService.FindInvitationsInTeam(ct, teamID1, &filter2)
	require.Nil(t, internalErr)
	require.Equal(t, 0, len(invitationsFound))
}

func TestInvitationService_FindInvitations(t *testing.T) {
	invitationRef, ok := prepareInvitationTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var invitationID1 uint64 = 12
	var invitationID2 uint64 = 13
	var invitationID3 uint64 = 14
	var receiverFirstName = "Test_FirstName"
	var receiverLastName = "Test_LastName"
	var receiverEmail = "test@test.com"
	var teamID1 uint64 = 5
	var teamID2 uint64 = 6
	var code1 = "Test code 1"
	var code2 = "Test code 2"
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := invitationRef.transactionFactory.BeginTx(ct, nil)
	require.Nil(t, err)

	defer tx.Rollback()

	now := time.Now().UTC()
	invitation1 := entity.Invitation{
		ID:                invitationID1,
		SenderUserID:      requesterUserID,
		ReceiverFirstName: &receiverFirstName,
		ReceiverLastName:  &receiverLastName,
		ReceiverEmail:     &receiverEmail,
		TeamID:            teamID1,
		ExpireAt:          now.Add(2 * time.Hour),
		Status:            entity.InvitationStatusPending,
		Code:              code2,
		CreatedAt:         now,
		UpdatedAt:         &now,
	}

	invitation2 := entity.Invitation{
		ID:                invitationID2,
		SenderUserID:      requesterUserID,
		ReceiverFirstName: &receiverFirstName,
		ReceiverLastName:  &receiverLastName,
		ReceiverEmail:     &receiverEmail,
		TeamID:            teamID2,
		ExpireAt:          now.Add(2 * time.Hour),
		Status:            entity.InvitationStatusPending,
		Code:              code1,
		CreatedAt:         now,
		UpdatedAt:         &now,
	}

	invitation3 := entity.Invitation{
		ID:                invitationID3,
		SenderUserID:      requesterUserID,
		ReceiverFirstName: &receiverFirstName,
		ReceiverLastName:  &receiverLastName,
		ReceiverEmail:     &receiverEmail,
		TeamID:            teamID2,
		ExpireAt:          now.Add(2 * time.Hour),
		Status:            entity.InvitationStatusPending,
		Code:              code2,
		CreatedAt:         now,
		UpdatedAt:         &now,
	}

	require.Nil(t, invitationRef.invitationDao.CreateInvitation(ct, tx, invitation1))
	require.Nil(t, invitationRef.invitationDao.CreateInvitation(ct, tx, invitation2))
	require.Nil(t, invitationRef.invitationDao.CreateInvitation(ct, tx, invitation3))

	filter1 := InvitationFilter{InvitationID: &invitationID2, Code: &code1}
	invitationsFound, internalErr := invitationRef.invitationService.FindInvitations(ct, &filter1)
	require.Nil(t, internalErr)
	require.Equal(t, 1, len(invitationsFound))
	require.True(t, areInvitationsEqual(invitation2, invitationsFound[0]))

	filter2 := InvitationFilter{InvitationID: &invitationID2, Code: &code2}
	invitationsFound, internalErr = invitationRef.invitationService.FindInvitations(ct, &filter2)
	require.Nil(t, internalErr)
	require.Equal(t, 0, len(invitationsFound))
}

func TestInvitationService_CreateInvitation(t *testing.T) {
	invitationRef, ok := prepareInvitationTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var receiverFirstName = "Test_FirstName"
	var receiverLastName = "Test_LastName"
	var receiverEmail = "test@test.com"
	var teamID uint64 = 5
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := invitationRef.transactionFactory.BeginTx(ct, nil)
	require.Nil(t, err)

	defer tx.Rollback()

	expiredAt := time.Now().Add(2 * time.Hour)
	input := CreateInvitationInput{
		ReceiverFirstName: &receiverFirstName,
		ReceiverLastName:  &receiverLastName,
		ReceiverEmail:     &receiverEmail,
		ExpireAt:          expiredAt,
	}

	invitation, err := invitationRef.invitationService.CreateInvitation(ct, teamID, input)
	require.Nil(t, err)
	require.Equal(t, receiverEmail, *(invitation.ReceiverEmail))
	require.Equal(t, receiverLastName, *(invitation.ReceiverLastName))
	require.Equal(t, receiverFirstName, *(invitation.ReceiverFirstName))
	require.Equal(t, teamID, invitation.TeamID)
	require.Equal(t, entity.InvitationStatusPending, invitation.Status)
	require.Equal(t, expiredAt, invitation.ExpireAt)
	require.NotNil(t, invitation.CreatedAt)
	require.Nil(t, invitation.UpdatedAt)

	invitationInMemory, err := invitationRef.invitationDao.FindInvitationByIDWithTx(ct, tx, invitation.ID)
	require.Nil(t, err)
	require.Equal(t, receiverEmail, *(invitationInMemory.ReceiverEmail))
	require.Equal(t, receiverLastName, *(invitationInMemory.ReceiverLastName))
	require.Equal(t, receiverFirstName, *(invitationInMemory.ReceiverFirstName))
	require.Equal(t, teamID, invitationInMemory.TeamID)
	require.Equal(t, entity.InvitationStatusPending, invitationInMemory.Status)
	require.Equal(t, expiredAt, invitationInMemory.ExpireAt)
	require.NotNil(t, invitationInMemory.CreatedAt)
	require.Nil(t, invitationInMemory.UpdatedAt)
}

func TestInvitationService_UpdateInvitation(t *testing.T) {
	invitationRef, ok := prepareInvitationTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var invitationID uint64 = 32
	var receiverFirstName = "Test_FirstName"
	var receiverLastName = "Test_LastName"
	var receiverEmail = "test@test.com"
	var code = "Test code"
	var teamID uint64 = 5
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := invitationRef.transactionFactory.BeginTx(ct, nil)
	require.Nil(t, err)

	defer tx.Rollback()

	now := time.Now()
	invitation := entity.Invitation{
		ID:                invitationID,
		SenderUserID:      requesterUserID,
		ReceiverFirstName: &receiverFirstName,
		ReceiverLastName:  &receiverLastName,
		ReceiverEmail:     &receiverEmail,
		TeamID:            teamID,
		ExpireAt:          now.Add(2 * time.Hour),
		Status:            entity.InvitationStatusPending,
		Code:              code,
		CreatedAt:         now,
		UpdatedAt:         nil,
	}

	err = invitationRef.invitationDao.CreateInvitation(ct, tx, invitation)
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
	require.Nil(t, err)

	invitation.ExpireAt = expiredAt
	invitation.ReceiverFirstName = &updatedFirstName
	invitation.ReceiverLastName = &updatedLastName
	require.True(t, areInvitationsEqual(invitation, updated))

	invitationInMemory, err := invitationRef.invitationDao.FindInvitationByIDWithTx(ct, tx, invitation.ID)
	require.Nil(t, err)
	require.True(t, areInvitationsEqual(invitation, invitationInMemory))
}

func TestInvitationService_DeleteInvitation(t *testing.T) {
	invitationRef, ok := prepareInvitationTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var invitationID uint64 = 32
	var receiverFirstName = "Test_FirstName"
	var receiverLastName = "Test_LastName"
	var receiverEmail = "test@test.com"
	var code = "Test code"
	var teamID uint64 = 5
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := invitationRef.transactionFactory.BeginTx(ct, nil)
	require.Nil(t, err)

	defer tx.Rollback()

	now := time.Now()
	invitation := entity.Invitation{
		ID:                invitationID,
		SenderUserID:      requesterUserID,
		ReceiverFirstName: &receiverFirstName,
		ReceiverLastName:  &receiverLastName,
		ReceiverEmail:     &receiverEmail,
		TeamID:            teamID,
		ExpireAt:          now.Add(2 * time.Hour),
		Status:            entity.InvitationStatusPending,
		Code:              code,
		CreatedAt:         now,
		UpdatedAt:         nil,
	}

	err = invitationRef.invitationDao.CreateInvitation(ct, tx, invitation)
	require.Nil(t, err)

	deleted, err := invitationRef.invitationService.DeleteInvitation(ct, invitation.ID)
	require.Nil(t, err)
	require.True(t, areInvitationsEqual(invitation, deleted))

	_, err = invitationRef.invitationDao.FindInvitationByIDWithTx(ct, tx, invitation.ID)
	require.NotNil(t, err)
	require.Equal(t, errs.NotFound, err.Code)
}

func TestInvitationService_AcceptInvitation(t *testing.T) {
	invitationRef, ok := prepareInvitationTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var invitationID uint64 = 32
	var receiverFirstName = "Test_FirstName"
	var receiverLastName = "Test_LastName"
	var receiverEmail = "test@test.com"
	var code = "Test code"
	var teamID uint64 = 5
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := invitationRef.transactionFactory.BeginTx(ct, nil)
	require.Nil(t, err)

	defer tx.Rollback()

	now := time.Now()
	invitation := entity.Invitation{
		ID:                invitationID,
		SenderUserID:      10,
		ReceiverFirstName: &receiverFirstName,
		ReceiverLastName:  &receiverLastName,
		ReceiverEmail:     &receiverEmail,
		TeamID:            teamID,
		ExpireAt:          now.Add(2 * time.Hour),
		Status:            entity.InvitationStatusPending,
		Code:              code,
		CreatedAt:         now,
		UpdatedAt:         nil,
	}

	err = invitationRef.invitationDao.CreateInvitation(ct, tx, invitation)
	require.Nil(t, err)

	accepted, err := invitationRef.invitationService.AcceptInvitation(ct, invitation.ID, invitation.Code)
	require.Nil(t, err)

	invitation.Status = entity.InvitationStatusAccepted
	invitation.ReceiverUserID = &requesterUserID
	require.True(t, areInvitationsEqual(invitation, accepted))

	invitationInMemory, err := invitationRef.invitationDao.FindInvitationByIDWithTx(ct, tx, invitation.ID)
	require.Nil(t, err)
	require.True(t, areInvitationsEqual(invitation, invitationInMemory))
}

func TestInvitationService_DeclineInvitation(t *testing.T) {
	invitationRef, ok := prepareInvitationTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var invitationID uint64 = 32
	var receiverFirstName = "Test_FirstName"
	var receiverLastName = "Test_LastName"
	var receiverEmail = "test@test.com"
	var code = "Test code"
	var teamID uint64 = 5
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := invitationRef.transactionFactory.BeginTx(ct, nil)
	require.Nil(t, err)

	defer tx.Rollback()

	now := time.Now()
	invitation := entity.Invitation{
		ID:                invitationID,
		SenderUserID:      10,
		ReceiverFirstName: &receiverFirstName,
		ReceiverLastName:  &receiverLastName,
		ReceiverEmail:     &receiverEmail,
		TeamID:            teamID,
		ExpireAt:          now.Add(2 * time.Hour),
		Status:            entity.InvitationStatusPending,
		Code:              code,
		CreatedAt:         now,
		UpdatedAt:         nil,
	}

	err = invitationRef.invitationDao.CreateInvitation(ct, tx, invitation)
	require.Nil(t, err)

	accepted, err := invitationRef.invitationService.DeclineInvitation(ct, invitation.ID, invitation.Code)

	invitation.Status = entity.InvitationStatusDeclined
	invitation.ReceiverUserID = &requesterUserID
	require.True(t, areInvitationsEqual(invitation, accepted))

	invitationInMemory, err := invitationRef.invitationDao.FindInvitationByIDWithTx(ct, tx, invitation.ID)
	require.Nil(t, err)
	require.True(t, areInvitationsEqual(invitation, invitationInMemory))
}

func areInvitationsEqual(one entity.Invitation, other entity.Invitation) bool {
	if one.ID != other.ID {
		return false
	}

	if one.TeamID != other.TeamID {
		return false
	}

	if one.Code != other.Code {
		return false
	}

	if one.ExpireAt != other.ExpireAt {
		return false
	}

	if one.CreatedAt != other.CreatedAt {
		return false
	}

	if one.ReceiverUserID == nil || other.ReceiverUserID == nil {
		if one.ReceiverUserID != nil || other.ReceiverUserID != nil {
			return false
		}
	} else if *one.ReceiverUserID != *other.ReceiverUserID {
		return false
	}

	if one.ReceiverLastName == nil || other.ReceiverLastName == nil {
		if one.ReceiverLastName != nil || other.ReceiverLastName != nil {
			return false
		}
	} else if *one.ReceiverLastName != *other.ReceiverLastName {
		return false
	}

	if one.ReceiverFirstName == nil || other.ReceiverFirstName == nil {
		if one.ReceiverFirstName != nil || other.ReceiverFirstName != nil {
			return false
		}
	} else if *one.ReceiverFirstName != *other.ReceiverFirstName {
		return false
	}

	if one.ReceiverEmail == nil || other.ReceiverEmail == nil {
		if one.ReceiverEmail != nil || other.ReceiverEmail != nil {
			return false
		}
	} else if *one.ReceiverEmail != *other.ReceiverEmail {
		return false
	}

	if one.SenderUserID != other.SenderUserID {
		return false
	}

	if one.Status != other.Status {
		return false
	}

	return true
}
