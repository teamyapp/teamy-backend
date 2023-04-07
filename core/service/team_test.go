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

type TeamTestRef struct {
	teamService                Team
	teamDaoV2                  daov2.Team
	teamMemberDaoV2            daov2.TeamMember
	teamFileUploadSessionDaoV2 daov2.TeamFileUploadSession
	transactionFactory         transaction.Factory
}

func prepareTeamTestRef(t *testing.T) (TeamTestRef, bool) {
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
		return TeamTestRef{}, false
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
		return TeamTestRef{}, false
	}

	authorizer := NewAuthorizer(logger, cloudClientRegistry)

	transactionFactory := transaction.NewFactory(nil)

	teamyBackendDB := dbtest.NewInMemoryDB()
	teamyBackendDB.CreateTable(daotestv2.TeamTableName)
	teamyBackendDB.CreateTable(daotestv2.TeamMemberTableName)
	teamyBackendDB.CreateTable(daotestv2.TeamFileUploadSessionTableName)
	teamyBackendDB.CreateTable(daotestv2.SprintTableName)

	teamMemberDao := daotest.NewTeamMember(teamyBackendDB)
	teamMemberDaoV2 := daotestv2.NewTeamMember(teamyBackendDB, transactionFactory)
	stateSyncer := realtime.NewStateSyncer(logger, teamMemberDao, teamMemberDaoV2)

	taskDaoV2 := daotestv2.NewTask(teamyBackendDB)
	sprintDao := daotest.NewSprint(teamyBackendDB)
	sprintDaoV2 := daotestv2.NewSprint(teamyBackendDB, transactionFactory)
	sprintParticipantDao := daotest.NewSprintParticipant(teamyBackendDB)
	sprintParticipantDaoV2 := daotestv2.NewSprintParticipant(teamyBackendDB, transactionFactory)
	teamDao := daotest.NewTeam(teamyBackendDB)
	teamDaoV2 := daotestv2.NewTeam(teamyBackendDB, transactionFactory)
	teamFileUploadSessionDaoV2 := daotestv2.NewTeamFileUploadSession(teamyBackendDB)

	teamService := NewTeam(
		logger,
		cloudTestKitConfig.WebAPIBaseURL,
		cloudClientRegistry,
		authorizer,
		stateSyncer,
		transactionFactory,
		taskDaoV2,
		sprintDao,
		sprintDaoV2,
		sprintParticipantDao,
		sprintParticipantDaoV2,
		teamDao,
		teamDaoV2,
		teamMemberDao,
		teamMemberDaoV2,
		teamFileUploadSessionDaoV2,
	)

	return TeamTestRef{
		teamService:                teamService,
		teamDaoV2:                  teamDaoV2,
		teamMemberDaoV2:            teamMemberDaoV2,
		teamFileUploadSessionDaoV2: teamFileUploadSessionDaoV2,
	}, true
}

func TestTeamService_FindTeamByID(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t)
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var teamID uint64 = 12
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := teamRef.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}

	defer tx.Rollback()

	var iconURL = "https://test"
	now := time.Now().UTC()
	team := entity.Team{
		ID:            teamID,
		Name:          "TeamName",
		IconURL:       &iconURL,
		CreatorUserID: requesterUserID,
		OwnerUserID:   requesterUserID,
		CreatedAt:     now,
		UpdatedAt:     &now,
	}

	// insert team into table
	if !assert.Nil(t, teamRef.teamDaoV2.CreateTeam(ct, tx, team)) {
		return
	}

	teamFound, internalErr := teamRef.teamService.FindTeamByID(ct, team.ID)
	if !assert.Nil(t, internalErr) {
		return
	}

	// verify return result
	assert.Equal(t, team.ID, teamFound.ID)
	assert.Equal(t, team.Name, teamFound.Name)
	assert.Equal(t, team.IconURL, teamFound.IconURL)
	assert.Equal(t, team.CreatorUserID, teamFound.CreatorUserID)
	assert.Equal(t, team.OwnerUserID, teamFound.OwnerUserID)
	assert.NotNil(t, team.CreatedAt, teamFound.CreatedAt)
	assert.NotNil(t, team.UpdatedAt, teamFound.UpdatedAt)
}

func TestTeamService_FindTeams(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t)
	if !ok {
		return
	}

	var requesterUserID1 uint64 = 20
	var requesterUserID2 uint64 = 32
	var teamID1 uint64 = 12
	var teamID2 uint64 = 2
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID1)

	tx, err := teamRef.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}

	defer tx.Rollback()

	var iconURL1 = "https://test1"
	var iconURL2 = "https://test2"
	now := time.Now().UTC()
	team1 := entity.Team{
		ID:            teamID1,
		Name:          "TeamName1",
		IconURL:       &iconURL1,
		CreatorUserID: requesterUserID1,
		OwnerUserID:   requesterUserID1,
		CreatedAt:     now,
		UpdatedAt:     &now,
	}
	team2 := entity.Team{
		ID:            teamID2,
		Name:          "TeamName2",
		IconURL:       &iconURL2,
		CreatorUserID: requesterUserID2,
		OwnerUserID:   requesterUserID2,
		CreatedAt:     now,
		UpdatedAt:     &now,
	}

	// insert teams into table
	if !assert.Nil(t, teamRef.teamDaoV2.CreateTeam(ct, tx, team1)) {
		return
	}
	if !assert.Nil(t, teamRef.teamDaoV2.CreateTeam(ct, tx, team2)) {
		return
	}

	teamFilter := TeamFilter{
		TeamID: &teamID2,
	}

	teamsFound, internalErr := teamRef.teamService.FindTeams(ct, &teamFilter)
	if !assert.Nil(t, internalErr) {
		return
	}

	// only team2 should be returned
	assert.Equal(t, len(teamsFound), 1)
	assert.Equal(t, team2.ID, teamsFound[0].ID)
	assert.Equal(t, team2.Name, teamsFound[0].Name)
	assert.Equal(t, team2.IconURL, teamsFound[0].IconURL)
	assert.Equal(t, team2.CreatorUserID, teamsFound[0].CreatorUserID)
	assert.Equal(t, team2.OwnerUserID, teamsFound[0].OwnerUserID)
	assert.NotNil(t, team2.CreatedAt, teamsFound[0].CreatedAt)
	assert.NotNil(t, team2.UpdatedAt, teamsFound[0].UpdatedAt)
}

func TestTeamService_FindTeamsForUser(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t)
	if !ok {
		return
	}

	var requesterUserID1 uint64 = 20
	var requesterUserID2 uint64 = 32
	var teamID1 uint64 = 12
	var teamID2 uint64 = 2
	var teamID3 uint64 = 4
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID1)

	tx, err := teamRef.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}

	defer tx.Rollback()

	var iconURL1 = "https://test1"
	var iconURL2 = "https://test2"
	var iconURL3 = "https://test3"
	now := time.Now().UTC()
	team1 := entity.Team{
		ID:            teamID1,
		Name:          "TeamName1",
		IconURL:       &iconURL1,
		CreatorUserID: requesterUserID1,
		OwnerUserID:   requesterUserID1,
		CreatedAt:     now,
		UpdatedAt:     &now,
	}
	team2 := entity.Team{
		ID:            teamID2,
		Name:          "TeamName2",
		IconURL:       &iconURL2,
		CreatorUserID: requesterUserID1,
		OwnerUserID:   requesterUserID1,
		CreatedAt:     now,
		UpdatedAt:     &now,
	}
	team3 := entity.Team{
		ID:            teamID3,
		Name:          "TeamName3",
		IconURL:       &iconURL3,
		CreatorUserID: requesterUserID1,
		OwnerUserID:   requesterUserID1,
		CreatedAt:     now,
		UpdatedAt:     &now,
	}

	teamMember1 := entity.TeamMember{TeamID: teamID1, UserID: requesterUserID1, CreatedAt: now}
	teamMember2 := entity.TeamMember{TeamID: teamID2, UserID: requesterUserID2, CreatedAt: now}
	teamMember3 := entity.TeamMember{TeamID: teamID3, UserID: requesterUserID1, CreatedAt: now}

	// insert teams and teamMembers into table
	if !assert.Nil(t, teamRef.teamDaoV2.CreateTeam(ct, tx, team1)) {
		return
	}

	if !assert.Nil(t, teamRef.teamDaoV2.CreateTeam(ct, tx, team2)) {
		return
	}

	if !assert.Nil(t, teamRef.teamDaoV2.CreateTeam(ct, tx, team3)) {
		return
	}
	if !assert.Nil(t, teamRef.teamMemberDaoV2.CreateTeamMember(ct, tx, teamMember1)) {
		return
	}

	if !assert.Nil(t, teamRef.teamMemberDaoV2.CreateTeamMember(ct, tx, teamMember2)) {
		return
	}

	if !assert.Nil(t, teamRef.teamMemberDaoV2.CreateTeamMember(ct, tx, teamMember3)) {
		return
	}

	teamFilter := TeamFilter{
		TeamID: &teamID3,
	}

	teamsFound, internalErr := teamRef.teamService.FindTeamsForUser(ct, requesterUserID1, &teamFilter)
	if !assert.Nil(t, internalErr) {
		return
	}

	// only team3 should be returned
	assert.Equal(t, 1, len(teamsFound))
	assert.Equal(t, team3.ID, teamsFound[0].ID)
	assert.Equal(t, team3.Name, teamsFound[0].Name)
	assert.Equal(t, team3.IconURL, teamsFound[0].IconURL)
	assert.Equal(t, team3.CreatorUserID, teamsFound[0].CreatorUserID)
	assert.Equal(t, team3.OwnerUserID, teamsFound[0].OwnerUserID)
	assert.NotNil(t, team3.CreatedAt, teamsFound[0].CreatedAt)
	assert.NotNil(t, team3.UpdatedAt, teamsFound[0].UpdatedAt)
}

func TestTeamService_CreateTeam(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t)
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)
	teamInput := CreateTeamInput{
		Name: "TeamName",
	}

	newTeam, internalErr := teamRef.teamService.CreateTeam(ct, teamInput)
	if !assert.Nil(t, internalErr) {
		return
	}

	// verify return result
	assert.Equal(t, requesterUserID, newTeam.CreatorUserID)
	assert.Equal(t, teamInput.Name, newTeam.Name)
	assert.Equal(t, requesterUserID, newTeam.OwnerUserID)
	assert.Nil(t, newTeam.IconURL)
	assert.NotNil(t, newTeam.CreatedAt)
	assert.Nil(t, newTeam.UpdatedAt)

	// verify in-memory DB
	teamInMemory, err := teamRef.teamDaoV2.FindTeamByID(ct, newTeam.ID)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, requesterUserID, teamInMemory.CreatorUserID)
	assert.Equal(t, teamInput.Name, teamInMemory.Name)
	assert.Equal(t, requesterUserID, teamInMemory.OwnerUserID)
	assert.Nil(t, teamInMemory.IconURL)
	assert.NotNil(t, teamInMemory.CreatedAt)
	assert.Nil(t, teamInMemory.UpdatedAt)
}

func TestTeamService_UpdateTeam(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t)
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var teamID uint64 = 12
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	var iconURL = "https://test1"
	now := time.Now().UTC()
	team := entity.Team{
		ID:            teamID,
		Name:          "TeamName",
		IconURL:       &iconURL,
		CreatorUserID: requesterUserID,
		OwnerUserID:   requesterUserID,
		CreatedAt:     now,
		UpdatedAt:     nil,
	}

	tx, err := teamRef.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}

	defer tx.Rollback()

	// insert team into table
	if !assert.Nil(t, teamRef.teamDaoV2.CreateTeam(ct, tx, team)) {
		return
	}

	updateTeamInput := UpdateTeamInput{Name: "UpdatedTeamName", OwnerUserID: 25}
	updatedTeam, internalErr := teamRef.teamService.UpdateTeam(ct, team.ID, updateTeamInput)
	if !assert.Nil(t, internalErr) {
		return
	}

	// verify return result
	assert.Equal(t, team.CreatorUserID, updatedTeam.CreatorUserID)
	assert.Equal(t, updateTeamInput.Name, updatedTeam.Name)
	assert.Equal(t, updateTeamInput.OwnerUserID, updatedTeam.OwnerUserID)
	assert.Equal(t, team.IconURL, updatedTeam.IconURL)
	assert.Equal(t, team.CreatedAt, updatedTeam.CreatedAt)
	assert.NotNil(t, updatedTeam.UpdatedAt)

	// verify in-memory DB
	teamInMemory, err := teamRef.teamDaoV2.FindTeamByID(ct, updatedTeam.ID)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, team.CreatorUserID, teamInMemory.CreatorUserID)
	assert.Equal(t, updateTeamInput.Name, teamInMemory.Name)
	assert.Equal(t, updateTeamInput.OwnerUserID, teamInMemory.OwnerUserID)
	assert.Equal(t, team.IconURL, teamInMemory.IconURL)
	assert.Equal(t, team.CreatedAt, teamInMemory.CreatedAt)
	assert.NotNil(t, teamInMemory.UpdatedAt)
}

func TestTeamService_DeleteTeam(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t)
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var teamID uint64 = 12
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	var iconURL = "https://test1"
	now := time.Now().UTC()
	team := entity.Team{
		ID:            teamID,
		Name:          "TeamName",
		IconURL:       &iconURL,
		CreatorUserID: requesterUserID,
		OwnerUserID:   requesterUserID,
		CreatedAt:     now,
		UpdatedAt:     &now,
	}

	tx, err := teamRef.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}
	defer tx.Rollback()

	// insert team into table
	if !assert.Nil(t, teamRef.teamDaoV2.CreateTeam(ct, tx, team)) {
		return
	}

	deletedTeam, internalErr := teamRef.teamService.DeleteTeam(ct, team.ID)
	if !assert.Nil(t, internalErr) {
		return
	}

	// verify return result
	assert.Equal(t, team.CreatorUserID, deletedTeam.CreatorUserID)
	assert.Equal(t, team.Name, deletedTeam.Name)
	assert.Equal(t, team.OwnerUserID, deletedTeam.OwnerUserID)
	assert.Equal(t, team.IconURL, deletedTeam.IconURL)
	assert.Equal(t, team.CreatedAt, deletedTeam.CreatedAt)
	assert.Equal(t, team.UpdatedAt, deletedTeam.UpdatedAt)

	// verify in-memory DB
	_, err = teamRef.teamDaoV2.FindTeamByID(ct, deletedTeam.ID)
	assert.NotNil(t, err)
	assert.Equal(t, err.Code, errs.NotFound)
}

func TestTeamService_CreateTeamIconUploadSession(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t)
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var teamID uint64 = 12
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)
	tx, err := teamRef.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}

	defer tx.Rollback()

	uploadSessionID, internalErr := teamRef.teamService.CreateTeamIconUploadSession(ct, teamID)
	if !assert.Nil(t, internalErr) {
		return
	}

	assert.Equal(t, uploadSessionID, uint64(1))

	// verify in-memory DB
	uploadSessionInMemory, err := teamRef.teamFileUploadSessionDaoV2.FindTeamFileUploadSessionByTeamIDWithTx(ct,
		tx,
		teamID,
		entity.IconTeamFileUploadSessionType,
		uploadSessionID)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, uploadSessionInMemory.FileUploadSessionID, uploadSessionID)
	assert.Equal(t, uploadSessionInMemory.TeamID, teamID)
	assert.Nil(t, uploadSessionInMemory.UpdatedAt)
	assert.Equal(t, uploadSessionInMemory.Type, entity.IconTeamFileUploadSessionType)
	assert.Equal(t, uploadSessionInMemory.IsCompleted, false)
}

func TestTeamService_FinishTeamIconUploadSession(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t)
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var teamID uint64 = 12
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	var iconURL = "https://test1"
	now := time.Now().UTC()
	team := entity.Team{
		ID:            teamID,
		Name:          "TeamName",
		IconURL:       &iconURL,
		CreatorUserID: requesterUserID,
		OwnerUserID:   requesterUserID,
		CreatedAt:     now,
		UpdatedAt:     &now,
	}

	tx, err := teamRef.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}

	defer tx.Rollback()

	// insert team into table
	if !assert.Nil(t, teamRef.teamDaoV2.CreateTeam(ct, tx, team)) {
		return
	}

	// create team upload session
	uploadSessionID, internalErr := teamRef.teamService.CreateTeamIconUploadSession(ct, teamID)
	if !assert.Nil(t, internalErr) {
		return
	}

	// finish upload session
	updatedTeam, err := teamRef.teamService.FinishTeamIconUploadSession(ct, team.ID, uploadSessionID)
	if !assert.Nil(t, err) {
		return
	}

	// verify returned team
	assert.Equal(t, team.ID, updatedTeam.ID)
	assert.NotEqual(t, team.IconURL, updatedTeam.IconURL)
	assert.NotEqual(t, team.UpdatedAt, updatedTeam.UpdatedAt)
	assert.Equal(t, team.Name, updatedTeam.Name)
	assert.Equal(t, team.CreatedAt, updatedTeam.CreatedAt)
	assert.Equal(t, team.OwnerUserID, updatedTeam.OwnerUserID)
	assert.Equal(t, team.CreatorUserID, updatedTeam.CreatorUserID)

	// verify in-memory DB
	uploadSessionInMemory, err := teamRef.teamFileUploadSessionDaoV2.FindTeamFileUploadSessionByTeamIDWithTx(ct,
		tx,
		teamID,
		entity.IconTeamFileUploadSessionType,
		uploadSessionID)
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, uploadSessionInMemory.FileUploadSessionID, uploadSessionID)
	assert.Equal(t, uploadSessionInMemory.TeamID, teamID)
	assert.NotNil(t, uploadSessionInMemory.UpdatedAt)
	assert.Equal(t, uploadSessionInMemory.Type, entity.IconTeamFileUploadSessionType)
	assert.Equal(t, uploadSessionInMemory.IsCompleted, true)
}

func TestTeamService_FindTeamMembers(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t)
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var teamID1 uint64 = 12
	var teamID2 uint64 = 24
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := teamRef.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}

	defer tx.Rollback()

	now := time.Now().UTC()
	teamMember1 := entity.TeamMember{
		TeamID:          teamID1,
		UserID:          requesterUserID,
		WeeklyBandwidth: timePerWeek,
		CreatedAt:       now,
		UpdatedAt:       &now,
	}
	teamMember2 := entity.TeamMember{
		TeamID:          teamID2,
		UserID:          requesterUserID,
		WeeklyBandwidth: timePerWeek,
		CreatedAt:       now,
		UpdatedAt:       &now,
	}

	// insert teams into table
	if !assert.Nil(t, teamRef.teamMemberDaoV2.CreateTeamMember(ct, tx, teamMember1)) {
		return
	}

	if !assert.Nil(t, teamRef.teamMemberDaoV2.CreateTeamMember(ct, tx, teamMember2)) {
		return
	}

	teamsFound, internalErr := teamRef.teamService.FindTeamMembers(ct, teamMember1.TeamID)
	if !assert.Nil(t, internalErr) {
		return
	}

	// verify return result
	assert.Equal(t, 1, len(teamsFound))
	assert.Equal(t, teamMember1.TeamID, teamsFound[0].TeamID)
	assert.Equal(t, teamMember1.UserID, teamsFound[0].UserID)
	assert.Equal(t, teamMember1.WeeklyBandwidth, teamsFound[0].WeeklyBandwidth)
	assert.Equal(t, teamMember1.CreatedAt, teamsFound[0].CreatedAt)
	assert.Equal(t, teamMember1.UpdatedAt, teamsFound[0].UpdatedAt)
}

func TestTeamService_AddMemberToTeam(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t)
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var teamID uint64 = 12
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := teamRef.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}
	defer tx.Rollback()

	teamMember, internalErr := teamRef.teamService.AddMemberToTeam(ct, teamID, requesterUserID)
	if !assert.Nil(t, internalErr) {
		return
	}

	// verify return result
	assert.Equal(t, teamMember.TeamID, teamID)
	assert.Equal(t, teamMember.UserID, requesterUserID)

	// verify in-memory DB
	teamMemberInMemory, internalErr := teamRef.teamMemberDaoV2.FindTeamMemberWithTx(ct, tx, teamID, requesterUserID)
	if !assert.Nil(t, internalErr) {
		return
	}

	assert.Equal(t, teamMemberInMemory.TeamID, teamID)
	assert.Equal(t, teamMemberInMemory.UserID, requesterUserID)
}

func TestTeamService_RemoveMemberFromTeam(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t)
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var teamID uint64 = 12
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := teamRef.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	teamMember := entity.TeamMember{
		TeamID:          teamID,
		UserID:          requesterUserID,
		WeeklyBandwidth: timePerWeek,
		CreatedAt:       now,
		UpdatedAt:       &now,
	}

	// insert teams into table
	if !assert.Nil(t, teamRef.teamMemberDaoV2.CreateTeamMember(ct, tx, teamMember)) {
		return
	}

	teamDeleted, internalErr := teamRef.teamService.RemoveMemberFromTeam(ct, teamMember.TeamID, teamMember.UserID)
	if !assert.Nil(t, internalErr) {
		return
	}

	// verify return result
	assert.Equal(t, teamDeleted.TeamID, teamMember.TeamID)
	assert.Equal(t, teamDeleted.UserID, teamMember.UserID)
	assert.Equal(t, teamDeleted.WeeklyBandwidth, teamMember.WeeklyBandwidth)
	assert.Equal(t, teamDeleted.CreatedAt, teamMember.CreatedAt)
	assert.Equal(t, teamDeleted.UpdatedAt, teamMember.UpdatedAt)

	// verify in-memory DB
	_, internalErr = teamRef.teamMemberDaoV2.FindTeamMemberWithTx(ct, tx, teamMember.TeamID, teamMember.UserID)
	assert.NotNil(t, internalErr)
	assert.Equal(t, internalErr.Code, errs.NotFound)
}

func TestTeamService_UpdateTeamMember(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t)
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var teamID uint64 = 12
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := teamRef.transactionFactory.BeginTx(ct, nil)
	if !assert.Nil(t, err) {
		return
	}

	defer tx.Rollback()

	now := time.Now().UTC()
	teamMember := entity.TeamMember{
		TeamID:          teamID,
		UserID:          requesterUserID,
		WeeklyBandwidth: timePerWeek,
		CreatedAt:       now,
		UpdatedAt:       &now,
	}

	// insert teams into table
	if !assert.Nil(t, teamRef.teamMemberDaoV2.CreateTeamMember(ct, tx, teamMember)) {
		return
	}

	updateInput := UpdateTeamMemberInput{
		UserID:          requesterUserID,
		WeeklyBandwidth: timePerWeek / 7,
	}
	teamUpdated, internalErr := teamRef.teamService.UpdateTeamMember(ct, teamMember.TeamID, updateInput)
	if !assert.Nil(t, internalErr) {
		return
	}

	// verify return result
	assert.Equal(t, teamUpdated.TeamID, teamMember.TeamID)
	assert.Equal(t, teamUpdated.UserID, teamMember.UserID)
	assert.Equal(t, teamUpdated.WeeklyBandwidth, updateInput.WeeklyBandwidth)
	assert.Equal(t, teamUpdated.CreatedAt, teamMember.CreatedAt)
	assert.NotEqual(t, teamUpdated.UpdatedAt, teamMember.UpdatedAt)

	// verify in-memory DB
	teamMemberInMemory, internalErr := teamRef.teamMemberDaoV2.FindTeamMemberWithTx(ct, tx, teamMember.TeamID, teamMember.UserID)
	if !assert.Nil(t, internalErr) {
		return
	}

	assert.Equal(t, teamMemberInMemory.TeamID, teamMember.TeamID)
	assert.Equal(t, teamMemberInMemory.UserID, teamMember.UserID)
	assert.Equal(t, teamMemberInMemory.WeeklyBandwidth, updateInput.WeeklyBandwidth)
	assert.Equal(t, teamMemberInMemory.CreatedAt, teamMember.CreatedAt)
	assert.NotEqual(t, teamMemberInMemory.UpdatedAt, teamMember.UpdatedAt)
}
