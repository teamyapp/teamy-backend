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
	"github.com/teamyapp/teamy-backend/core/dao/daotest"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/daov2/daotestv2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type InvitationTestRef struct {
	invitationService  Invitation
	invitationDaoV2    daov2.Invitation
	transactionFactory transaction.Factory
}

func prepareInvitationTestRef(t *testing.T) (InvitationTestRef, bool) {
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

	teamMemberDao := daotest.NewTeamMember(teamyBackendDB)
	teamMemberDaoV2 := daotestv2.NewTeamMember(teamyBackendDB, transactionFactory)
	stateSyncer := realtime.NewStateSyncer(logger, teamMemberDao, teamMemberDaoV2)

	invitationDao := daotest.NewInvitation(teamyBackendDB)
	invitationDaoV2 := daotestv2.NewInvitation(teamyBackendDB, transactionFactory)
	sprintParticipantDao := daotest.NewSprintParticipant(teamyBackendDB)
	sprintParticipantDaoV2 := daotestv2.NewSprintParticipant(teamyBackendDB, transactionFactory)
	sprintDao := daotest.NewSprint(teamyBackendDB)
	sprintDaoV2 := daotestv2.NewSprint(teamyBackendDB, transactionFactory)
	invitationService := NewInvitation(
		logger,
		cloudClientRegistry,
		authorizer,
		stateSyncer,
		transactionFactory,
		invitationDao,
		invitationDaoV2,
		teamMemberDao,
		teamMemberDaoV2,
		sprintParticipantDao,
		sprintParticipantDaoV2,
		sprintDao,
		sprintDaoV2,
	)
	return InvitationTestRef{
		invitationService:  invitationService,
		invitationDaoV2:    invitationDaoV2,
		transactionFactory: transactionFactory,
	}, true
}

func TestInvitationService_FindInvitationsInTeam(t *testing.T) {
	invitationRef, ok := prepareInvitationTestRef(t)
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
	invitationFound, internalErr := invitationRef.invitationService.FindInvitationsInTeam(ct, teamID2, &filter1)
	if !assert.Nil(t, internalErr) {
		return
	}

	// verify return result
	assert.Equal(t, 1, len(invitationFound))
	assert.Equal(t, invitation2.ID, invitationFound[0].ID)
	assert.Equal(t, *(invitation2.ReceiverEmail), *(invitationFound[0].ReceiverEmail))
	assert.Equal(t, *(invitation2.ReceiverLastName), *(invitationFound[0].ReceiverLastName))
	assert.Equal(t, *(invitation2.ReceiverFirstName), *(invitationFound[0].ReceiverFirstName))
	assert.Equal(t, invitation2.ReceiverUserID, invitationFound[0].ReceiverUserID)
	assert.Equal(t, invitation2.TeamID, invitationFound[0].TeamID)
	assert.Equal(t, invitation2.Status, invitationFound[0].Status)
	assert.Equal(t, invitation2.ExpireAt, invitationFound[0].ExpireAt)
	assert.Equal(t, invitation2.CreatedAt, invitationFound[0].CreatedAt)
	assert.Equal(t, *(invitation2.UpdatedAt), *(invitationFound[0].UpdatedAt))

	filter2 := InvitationFilter{InvitationID: &invitationID2, Code: &code2}
	invitationsFound, internalErr = invitationRef.invitationService.FindInvitationsInTeam(ct, teamID1, &filter2)
	if !assert.Nil(t, internalErr) {
		return
	}

	assert.Equal(t, 0, len(invitationsFound))
}

func TestInvitationService_FindInvitations(t *testing.T) {
	invitationRef, ok := prepareInvitationTestRef(t)
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
	invitationFound, internalErr := invitationRef.invitationService.FindInvitations(ct, &filter1)
	if !assert.Nil(t, internalErr) {
		return
	}

	// verify return result
	assert.Equal(t, 1, len(invitationFound))
	assert.Equal(t, invitation2.ID, invitationFound[0].ID)
	assert.Equal(t, *(invitation2.ReceiverEmail), *(invitationFound[0].ReceiverEmail))
	assert.Equal(t, *(invitation2.ReceiverLastName), *(invitationFound[0].ReceiverLastName))
	assert.Equal(t, *(invitation2.ReceiverFirstName), *(invitationFound[0].ReceiverFirstName))
	assert.Equal(t, invitation2.ReceiverUserID, invitationFound[0].ReceiverUserID)
	assert.Equal(t, invitation2.TeamID, invitationFound[0].TeamID)
	assert.Equal(t, invitation2.Status, invitationFound[0].Status)
	assert.Equal(t, invitation2.ExpireAt, invitationFound[0].ExpireAt)
	assert.Equal(t, invitation2.CreatedAt, invitationFound[0].CreatedAt)
	assert.Equal(t, *(invitation2.UpdatedAt), *(invitationFound[0].UpdatedAt))

	filter2 := InvitationFilter{InvitationID: &invitationID2, Code: &code2}
	invitationFound, internalErr = invitationRef.invitationService.FindInvitations(ct, &filter2)
	if !assert.Nil(t, internalErr) {
		return
	}

	assert.Equal(t, 0, len(invitationFound))
}

func TestInvitationService_CreateInvitation(t *testing.T) {
	invitationRef, ok := prepareInvitationTestRef(t)
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
	invitationRef, ok := prepareInvitationTestRef(t)
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
	assert.Equal(t, *(invitation.ReceiverEmail), *(updated.ReceiverEmail))
	assert.Equal(t, updatedLastName, *(updated.ReceiverLastName))
	assert.Equal(t, updatedFirstName, *(updated.ReceiverFirstName))
	assert.Equal(t, invitation.TeamID, updated.TeamID)
	assert.Equal(t, invitation.Status, updated.Status)
	assert.Equal(t, expiredAt, updated.ExpireAt)
	assert.Equal(t, invitation.Code, updated.Code)
	assert.Equal(t, invitation.SenderUserID, updated.SenderUserID)
	assert.Equal(t, invitation.CreatedAt, updated.CreatedAt)
	assert.NotNil(t, updated.UpdatedAt)

	// verify in-memory DB
	invitationInMemory, err := invitationRef.invitationDaoV2.FindInvitationByIDWithTx(ct, tx, invitation.ID)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, *(invitation.ReceiverEmail), *(invitationInMemory.ReceiverEmail))
	assert.Equal(t, updatedLastName, *(invitationInMemory.ReceiverLastName))
	assert.Equal(t, updatedFirstName, *(invitationInMemory.ReceiverFirstName))
	assert.Equal(t, invitation.TeamID, invitationInMemory.TeamID)
	assert.Equal(t, invitation.Status, invitationInMemory.Status)
	assert.Equal(t, expiredAt, invitationInMemory.ExpireAt)
	assert.Equal(t, invitation.Code, invitationInMemory.Code)
	assert.Equal(t, invitation.SenderUserID, invitationInMemory.SenderUserID)
	assert.Equal(t, invitation.CreatedAt, invitationInMemory.CreatedAt)
	assert.NotNil(t, invitationInMemory.UpdatedAt)
}

func TestInvitationService_DeleteInvitation(t *testing.T) {
	invitationRef, ok := prepareInvitationTestRef(t)
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
	assert.Equal(t, *(invitation.ReceiverEmail), *(deleted.ReceiverEmail))
	assert.Equal(t, *(invitation.ReceiverLastName), *(deleted.ReceiverLastName))
	assert.Equal(t, *(invitation.ReceiverFirstName), *(deleted.ReceiverFirstName))
	assert.Equal(t, invitation.TeamID, deleted.TeamID)
	assert.Equal(t, invitation.Status, deleted.Status)
	assert.Equal(t, invitation.ExpireAt, deleted.ExpireAt)
	assert.Equal(t, invitation.Code, deleted.Code)
	assert.Equal(t, invitation.SenderUserID, deleted.SenderUserID)
	assert.Equal(t, invitation.CreatedAt, deleted.CreatedAt)

	// verify in-memory DB
	_, err = invitationRef.invitationDaoV2.FindInvitationByIDWithTx(ct, tx, invitation.ID)
	if !assert.NotNil(t, err) {
		return
	}
	assert.Equal(t, errs.NotFound, err.Code)
}

func TestInvitationService_AcceptInvitation(t *testing.T) {
	invitationRef, ok := prepareInvitationTestRef(t)
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

	accepted, err := invitationRef.invitationService.AcceptInvitation(ct, invitation.ID, invitation.Code)
	if !assert.Nil(t, err) {
		return
	}

	// verify return result
	assert.Equal(t, *(invitation.ReceiverEmail), *(accepted.ReceiverEmail))
	assert.Equal(t, *(invitation.ReceiverLastName), *(accepted.ReceiverLastName))
	assert.Equal(t, *(invitation.ReceiverFirstName), *(accepted.ReceiverFirstName))
	assert.Equal(t, requesterUserID, *(accepted.ReceiverUserID))
	assert.Equal(t, invitation.TeamID, accepted.TeamID)
	assert.Equal(t, entity.InvitationStatusAccepted, accepted.Status)
	assert.Equal(t, invitation.ExpireAt, accepted.ExpireAt)
	assert.Equal(t, invitation.Code, accepted.Code)
	assert.Equal(t, invitation.SenderUserID, accepted.SenderUserID)
	assert.Equal(t, invitation.CreatedAt, accepted.CreatedAt)
	assert.NotNil(t, accepted.UpdatedAt)

	// verify in-memory DB
	invitationInMemory, err := invitationRef.invitationDaoV2.FindInvitationByIDWithTx(ct, tx, invitation.ID)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, *(invitation.ReceiverEmail), *(invitationInMemory.ReceiverEmail))
	assert.Equal(t, *(invitation.ReceiverLastName), *(invitationInMemory.ReceiverLastName))
	assert.Equal(t, *(invitation.ReceiverFirstName), *(invitationInMemory.ReceiverFirstName))
	assert.Equal(t, requesterUserID, *(invitationInMemory.ReceiverUserID))
	assert.Equal(t, invitation.TeamID, invitationInMemory.TeamID)
	assert.Equal(t, entity.InvitationStatusAccepted, invitationInMemory.Status)
	assert.Equal(t, invitation.ExpireAt, invitationInMemory.ExpireAt)
	assert.Equal(t, invitation.Code, invitationInMemory.Code)
	assert.Equal(t, invitation.SenderUserID, invitationInMemory.SenderUserID)
	assert.Equal(t, invitation.CreatedAt, invitationInMemory.CreatedAt)
	assert.NotNil(t, invitationInMemory.UpdatedAt)
}

func TestInvitationService_DeclineInvitation(t *testing.T) {
	invitationRef, ok := prepareInvitationTestRef(t)
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

	accepted, err := invitationRef.invitationService.DeclineInvitation(ct, invitation.ID, invitation.Code)

	// verify return result
	assert.Equal(t, *(invitation.ReceiverEmail), *(accepted.ReceiverEmail))
	assert.Equal(t, *(invitation.ReceiverLastName), *(accepted.ReceiverLastName))
	assert.Equal(t, *(invitation.ReceiverFirstName), *(accepted.ReceiverFirstName))
	assert.Equal(t, requesterUserID, *(accepted.ReceiverUserID))
	assert.Equal(t, invitation.TeamID, accepted.TeamID)
	assert.Equal(t, entity.InvitationStatusDeclined, accepted.Status)
	assert.Equal(t, invitation.ExpireAt, accepted.ExpireAt)
	assert.Equal(t, invitation.Code, accepted.Code)
	assert.Equal(t, invitation.SenderUserID, accepted.SenderUserID)
	assert.Equal(t, invitation.CreatedAt, accepted.CreatedAt)
	assert.NotNil(t, accepted.UpdatedAt)

	// verify in-memory DB
	invitationInMemory, err := invitationRef.invitationDaoV2.FindInvitationByIDWithTx(ct, tx, invitation.ID)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, *(invitation.ReceiverEmail), *(invitationInMemory.ReceiverEmail))
	assert.Equal(t, *(invitation.ReceiverLastName), *(invitationInMemory.ReceiverLastName))
	assert.Equal(t, *(invitation.ReceiverFirstName), *(invitationInMemory.ReceiverFirstName))
	assert.Equal(t, requesterUserID, *(invitationInMemory.ReceiverUserID))
	assert.Equal(t, invitation.TeamID, invitationInMemory.TeamID)
	assert.Equal(t, entity.InvitationStatusDeclined, invitationInMemory.Status)
	assert.Equal(t, invitation.ExpireAt, invitationInMemory.ExpireAt)
	assert.Equal(t, invitation.Code, invitationInMemory.Code)
	assert.Equal(t, invitation.SenderUserID, invitationInMemory.SenderUserID)
	assert.Equal(t, invitation.CreatedAt, invitationInMemory.CreatedAt)
	assert.NotNil(t, invitationInMemory.UpdatedAt)
}
