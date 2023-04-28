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
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/daov2/daotestv2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type InvitationTestRef struct {
	invitationService  Invitation
	invitationDaoV2    daov2.Invitation
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
	if !assert.Nil(t, internalErr) {
		return InvitationTestRef{}, false
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
		return InvitationTestRef{}, false
	}

	authorizer := NewAuthorizer(logger, cloudClientRegistry)

	transactionFactory := transaction.NewFactory(nil)

	teamyBackendDB := dbtest.NewInMemoryDB()
	teamyBackendDB.CreateTable(daotestv2.InvitationTableName)
	teamyBackendDB.CreateTable(daotestv2.TeamMemberTableName)
	teamyBackendDB.CreateTable(daotestv2.SprintTableName)

	teamMemberDaoV2 := daotestv2.NewTeamMember(teamyBackendDB, transactionFactory)
	stateSyncer := realtime.NewStateSyncer(logger, teamMemberDaoV2)

	invitationDaoV2 := daotestv2.NewInvitation(teamyBackendDB, transactionFactory)
	sprintParticipantDaoV2 := daotestv2.NewSprintParticipant(teamyBackendDB, transactionFactory)
	sprintDaoV2 := daotestv2.NewSprint(teamyBackendDB, transactionFactory)
	invitationService := NewInvitation(
		logger,
		cloudClientRegistry,
		authorizer,
		toggles,
		stateSyncer,
		transactionFactory,
		invitationDaoV2,
		teamMemberDaoV2,
		sprintParticipantDaoV2,
		sprintDaoV2,
	)
	return InvitationTestRef{
		invitationService:  invitationService,
		invitationDaoV2:    invitationDaoV2,
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
	if !assert.Nil(t, err) {
		return
	}

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

	// insert team into table
	if !assert.Nil(t, invitationRef.invitationDaoV2.CreateInvitation(ct, tx, invitation1)) {
		return
	}

	if !assert.Nil(t, invitationRef.invitationDaoV2.CreateInvitation(ct, tx, invitation2)) {
		return
	}

	if !assert.Nil(t, invitationRef.invitationDaoV2.CreateInvitation(ct, tx, invitation3)) {
		return
	}

	filter1 := InvitationFilter{InvitationID: &invitationID2, Code: &code1}
	invitationsFound, internalErr := invitationRef.invitationService.FindInvitationsInTeam(ct, teamID2, &filter1)
	if !assert.Nil(t, internalErr) {
		return
	}

	// verify return result
	assert.Equal(t, 1, len(invitationsFound))

	assert.Equal(t, invitation2.ID, invitationsFound[0].ID)
	assert.True(t, areInvitationsEqual(invitation2, invitationsFound[0]))

	filter2 := InvitationFilter{InvitationID: &invitationID2, Code: &code2}
	invitationsFound, internalErr = invitationRef.invitationService.FindInvitationsInTeam(ct, teamID1, &filter2)
	if !assert.Nil(t, internalErr) {
		return
	}

	assert.Equal(t, 0, len(invitationsFound))
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
	if !assert.Nil(t, err) {
		return
	}

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

	// insert team into table
	if !assert.Nil(t, invitationRef.invitationDaoV2.CreateInvitation(ct, tx, invitation1)) {
		return
	}

	if !assert.Nil(t, invitationRef.invitationDaoV2.CreateInvitation(ct, tx, invitation2)) {
		return
	}

	if !assert.Nil(t, invitationRef.invitationDaoV2.CreateInvitation(ct, tx, invitation3)) {
		return
	}

	filter1 := InvitationFilter{InvitationID: &invitationID2, Code: &code1}
	invitationsFound, internalErr := invitationRef.invitationService.FindInvitations(ct, &filter1)
	if !assert.Nil(t, internalErr) {
		return
	}

	// verify return result
	assert.Equal(t, 1, len(invitationsFound))
	assert.True(t, areInvitationsEqual(invitation2, invitationsFound[0]))

	filter2 := InvitationFilter{InvitationID: &invitationID2, Code: &code2}
	invitationsFound, internalErr = invitationRef.invitationService.FindInvitations(ct, &filter2)
	if !assert.Nil(t, internalErr) {
		return
	}

	assert.Equal(t, 0, len(invitationsFound))
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
	if !assert.Nil(t, err) {
		return
	}

	defer tx.Rollback()

	expiredAt := time.Now().Add(2 * time.Hour)
	input := CreateInvitationInput{
		ReceiverFirstName: &receiverFirstName,
		ReceiverLastName:  &receiverLastName,
		ReceiverEmail:     &receiverEmail,
		ExpireAt:          expiredAt,
	}

	invitation, err := invitationRef.invitationService.CreateInvitation(ct, teamID, input)
	if !assert.Nil(t, err) {
		return
	}

	// verify return result
	assert.Equal(t, receiverEmail, *(invitation.ReceiverEmail))
	assert.Equal(t, receiverLastName, *(invitation.ReceiverLastName))
	assert.Equal(t, receiverFirstName, *(invitation.ReceiverFirstName))
	assert.Equal(t, teamID, invitation.TeamID)
	assert.Equal(t, entity.InvitationStatusPending, invitation.Status)
	assert.Equal(t, expiredAt, invitation.ExpireAt)
	assert.NotNil(t, invitation.CreatedAt)
	assert.Nil(t, invitation.UpdatedAt)

	// verify in-memory DB
	invitationInMemory, err := invitationRef.invitationDaoV2.FindInvitationByIDWithTx(ct, tx, invitation.ID)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, receiverEmail, *(invitationInMemory.ReceiverEmail))
	assert.Equal(t, receiverLastName, *(invitationInMemory.ReceiverLastName))
	assert.Equal(t, receiverFirstName, *(invitationInMemory.ReceiverFirstName))
	assert.Equal(t, teamID, invitationInMemory.TeamID)
	assert.Equal(t, entity.InvitationStatusPending, invitationInMemory.Status)
	assert.Equal(t, expiredAt, invitationInMemory.ExpireAt)
	assert.NotNil(t, invitationInMemory.CreatedAt)
	assert.Nil(t, invitationInMemory.UpdatedAt)
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
	if !assert.Nil(t, err) {
		return
	}

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

	// insert team into table
	err = invitationRef.invitationDaoV2.CreateInvitation(ct, tx, invitation)
	if !assert.Nil(t, err) {
		return
	}

	expiredAt := now.Add(3 * time.Hour)
	updatedFirstName := "Updated_FirstName"
	updatedLastName := "Updated_LastName"
	input := UpdateInvitationInput{
		ExpireAt:          expiredAt,
		ReceiverFirstName: &updatedFirstName,
		ReceiverLastName:  &updatedLastName,
	}
	updated, err := invitationRef.invitationService.UpdateInvitation(ct, invitation.ID, input)
	if !assert.Nil(t, err) {
		return
	}

	// verify return result
	invitation.ExpireAt = expiredAt
	invitation.ReceiverFirstName = &updatedFirstName
	invitation.ReceiverLastName = &updatedLastName
	assert.True(t, areInvitationsEqual(invitation, updated))

	// verify in-memory DB
	invitationInMemory, err := invitationRef.invitationDaoV2.FindInvitationByIDWithTx(ct, tx, invitation.ID)
	if !assert.Nil(t, err) {
		return
	}

	assert.True(t, areInvitationsEqual(invitation, invitationInMemory))
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
	if !assert.Nil(t, err) {
		return
	}

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

	// insert team into table
	err = invitationRef.invitationDaoV2.CreateInvitation(ct, tx, invitation)
	if !assert.Nil(t, err) {
		return
	}

	deleted, err := invitationRef.invitationService.DeleteInvitation(ct, invitation.ID)
	if !assert.Nil(t, err) {
		return
	}

	// verify return result
	assert.True(t, areInvitationsEqual(invitation, deleted))

	// verify in-memory DB
	_, err = invitationRef.invitationDaoV2.FindInvitationByIDWithTx(ct, tx, invitation.ID)
	if !assert.NotNil(t, err) {
		return
	}

	assert.Equal(t, errs.NotFound, err.Code)
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
	if !assert.Nil(t, err) {
		return
	}

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

	// insert team into table
	err = invitationRef.invitationDaoV2.CreateInvitation(ct, tx, invitation)
	if !assert.Nil(t, err) {
		return
	}

	accepted, err := invitationRef.invitationService.AcceptInvitation(ct, invitation.ID, invitation.Code)
	if !assert.Nil(t, err) {
		return
	}

	// verify return result
	invitation.Status = entity.InvitationStatusAccepted
	invitation.ReceiverUserID = &requesterUserID
	assert.True(t, areInvitationsEqual(invitation, accepted))

	// verify in-memory DB
	invitationInMemory, err := invitationRef.invitationDaoV2.FindInvitationByIDWithTx(ct, tx, invitation.ID)
	if !assert.Nil(t, err) {
		return
	}

	assert.True(t, areInvitationsEqual(invitation, invitationInMemory))
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
	if !assert.Nil(t, err) {
		return
	}

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

	// insert team into table
	err = invitationRef.invitationDaoV2.CreateInvitation(ct, tx, invitation)
	if !assert.Nil(t, err) {
		return
	}

	accepted, err := invitationRef.invitationService.DeclineInvitation(ct, invitation.ID, invitation.Code)

	// verify return result
	invitation.Status = entity.InvitationStatusDeclined
	invitation.ReceiverUserID = &requesterUserID
	assert.True(t, areInvitationsEqual(invitation, accepted))

	// verify in-memory DB
	invitationInMemory, err := invitationRef.invitationDaoV2.FindInvitationByIDWithTx(ct, tx, invitation.ID)
	if !assert.Nil(t, err) {
		return
	}

	assert.True(t, areInvitationsEqual(invitation, invitationInMemory))
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
