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
	"github.com/teamyapp/teamy-backend/core/repository"
	"github.com/teamyapp/teamy-backend/core/service/servicetest"
)

type TeamTestRef struct {
	transactionFactory             transaction.Factory
	teamService                    Team
	teamDao                        dao.Team
	teamMemberDao                  dao.TeamMember
	teamFileUploadSessionDao       dao.TeamFileUploadSession
	teamMemberGroupDao             dao.TeamMemberGroup
	teamMemberGroupUserRelationDao dao.TeamMemberGroupUserRelation
	cloudTestKit                   testkit.TestKit
}

func TestTeamService_FindTeamByID(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var teamID uint64 = 12
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := teamRef.transactionFactory.BeginTx(ct, nil)
	require.Nil(t, err)

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

	require.Nil(t, teamRef.teamDao.CreateTeam(ct, tx, team))

	teamFound, internalErr := teamRef.teamService.FindTeamByID(ct, team.ID)
	require.Nil(t, internalErr)
	require.Equal(t, team.ID, teamFound.ID)
	require.Equal(t, team.Name, teamFound.Name)
	require.Equal(t, team.IconURL, teamFound.IconURL)
	require.Equal(t, team.CreatorUserID, teamFound.CreatorUserID)
	require.Equal(t, team.OwnerUserID, teamFound.OwnerUserID)
	require.NotNil(t, team.CreatedAt, teamFound.CreatedAt)
	require.NotNil(t, team.UpdatedAt, teamFound.UpdatedAt)
}

func TestTeamService_FindTeams(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
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
	require.Nil(t, err)

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

	require.Nil(t, teamRef.teamDao.CreateTeam(ct, tx, team1))
	require.Nil(t, teamRef.teamDao.CreateTeam(ct, tx, team2))

	teamFilter := TeamFilter{
		TeamID: &teamID2,
	}

	teamsFound, internalErr := teamRef.teamService.FindTeams(ct, &teamFilter)
	require.Nil(t, internalErr)
	// only team2 should be returned
	require.Equal(t, len(teamsFound), 1)
	require.Equal(t, team2.ID, teamsFound[0].ID)
	require.Equal(t, team2.Name, teamsFound[0].Name)
	require.Equal(t, team2.IconURL, teamsFound[0].IconURL)
	require.Equal(t, team2.CreatorUserID, teamsFound[0].CreatorUserID)
	require.Equal(t, team2.OwnerUserID, teamsFound[0].OwnerUserID)
	require.NotNil(t, team2.CreatedAt, teamsFound[0].CreatedAt)
	require.NotNil(t, team2.UpdatedAt, teamsFound[0].UpdatedAt)
}

func TestTeamService_FindTeamsForUser(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
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
	require.Nil(t, err)

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

	require.Nil(t, teamRef.teamDao.CreateTeam(ct, tx, team1))
	require.Nil(t, teamRef.teamDao.CreateTeam(ct, tx, team2))
	require.Nil(t, teamRef.teamDao.CreateTeam(ct, tx, team3))
	require.Nil(t, teamRef.teamMemberDao.CreateTeamMember(ct, tx, teamMember1))
	require.Nil(t, teamRef.teamMemberDao.CreateTeamMember(ct, tx, teamMember2))
	require.Nil(t, teamRef.teamMemberDao.CreateTeamMember(ct, tx, teamMember3))

	teamFilter := TeamFilter{
		TeamID: &teamID3,
	}

	teamsFound, internalErr := teamRef.teamService.FindTeamsForUser(ct, requesterUserID1, &teamFilter)
	require.Nil(t, internalErr)
	// only team3 should be returned
	require.Equal(t, 1, len(teamsFound))
	require.Equal(t, team3.ID, teamsFound[0].ID)
	require.Equal(t, team3.Name, teamsFound[0].Name)
	require.Equal(t, team3.IconURL, teamsFound[0].IconURL)
	require.Equal(t, team3.CreatorUserID, teamsFound[0].CreatorUserID)
	require.Equal(t, team3.OwnerUserID, teamsFound[0].OwnerUserID)
	require.NotNil(t, team3.CreatedAt, teamsFound[0].CreatedAt)
	require.NotNil(t, team3.UpdatedAt, teamsFound[0].UpdatedAt)
}

func TestTeamService_CreateTeam(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t, feature.Toggles{
		EnableAuthorization: true,
	})
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
	require.Nil(t, internalErr)

	require.Equal(t, requesterUserID, newTeam.CreatorUserID)
	require.Equal(t, teamInput.Name, newTeam.Name)
	require.Equal(t, requesterUserID, newTeam.OwnerUserID)
	require.Nil(t, newTeam.IconURL)
	require.NotNil(t, newTeam.CreatedAt)
	require.Nil(t, newTeam.UpdatedAt)

	teamInMemory, err := teamRef.teamDao.FindTeamByID(ct, newTeam.ID)
	require.Nil(t, err)

	require.Equal(t, requesterUserID, teamInMemory.CreatorUserID)
	require.Equal(t, teamInput.Name, teamInMemory.Name)
	require.Equal(t, requesterUserID, teamInMemory.OwnerUserID)
	require.Nil(t, teamInMemory.IconURL)
	require.NotNil(t, teamInMemory.CreatedAt)
	require.Nil(t, teamInMemory.UpdatedAt)
}

func TestTeamService_UpdateTeam(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
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
	require.Nil(t, err)

	defer tx.Rollback()

	require.Nil(t, teamRef.teamDao.CreateTeam(ct, tx, team))

	updateTeamInput := UpdateTeamInput{Name: "UpdatedTeamName", OwnerUserID: 25}
	updatedTeam, internalErr := teamRef.teamService.UpdateTeam(ct, team.ID, updateTeamInput)
	require.Nil(t, internalErr)
	require.Equal(t, team.CreatorUserID, updatedTeam.CreatorUserID)
	require.Equal(t, updateTeamInput.Name, updatedTeam.Name)
	require.Equal(t, updateTeamInput.OwnerUserID, updatedTeam.OwnerUserID)
	require.Equal(t, team.IconURL, updatedTeam.IconURL)
	require.Equal(t, team.CreatedAt, updatedTeam.CreatedAt)
	require.NotNil(t, updatedTeam.UpdatedAt)

	teamInMemory, err := teamRef.teamDao.FindTeamByID(ct, updatedTeam.ID)
	require.Nil(t, err)
	require.Equal(t, team.CreatorUserID, teamInMemory.CreatorUserID)
	require.Equal(t, updateTeamInput.Name, teamInMemory.Name)
	require.Equal(t, updateTeamInput.OwnerUserID, teamInMemory.OwnerUserID)
	require.Equal(t, team.IconURL, teamInMemory.IconURL)
	require.Equal(t, team.CreatedAt, teamInMemory.CreatedAt)
	require.NotNil(t, teamInMemory.UpdatedAt)
}

func TestTeamService_DeleteTeam(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
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
	require.Nil(t, err)

	defer tx.Rollback()

	require.Nil(t, teamRef.teamDao.CreateTeam(ct, tx, team))

	deletedTeam, internalErr := teamRef.teamService.DeleteTeam(ct, team.ID)
	require.Nil(t, internalErr)
	require.Equal(t, team.CreatorUserID, deletedTeam.CreatorUserID)
	require.Equal(t, team.Name, deletedTeam.Name)
	require.Equal(t, team.OwnerUserID, deletedTeam.OwnerUserID)
	require.Equal(t, team.IconURL, deletedTeam.IconURL)
	require.Equal(t, team.CreatedAt, deletedTeam.CreatedAt)
	require.Equal(t, team.UpdatedAt, deletedTeam.UpdatedAt)

	_, err = teamRef.teamDao.FindTeamByID(ct, deletedTeam.ID)
	require.NotNil(t, err)
	require.Equal(t, err.Code, errs.NotFound)
}

func TestTeamService_CreateTeamIconUploadSession(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var teamID uint64 = 12
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)
	tx, err := teamRef.transactionFactory.BeginTx(ct, nil)
	require.Nil(t, err)

	defer tx.Rollback()

	uploadSessionID, internalErr := teamRef.teamService.CreateTeamIconUploadSession(ct, teamID)
	require.Nil(t, internalErr)
	require.Equal(t, uploadSessionID, uint64(1))

	uploadSessionInMemory, err := teamRef.teamFileUploadSessionDao.FindTeamFileUploadSessionByTeamIDWithTx(ct,
		tx,
		teamID,
		entity.IconTeamFileUploadSessionType,
		uploadSessionID)
	require.Nil(t, err)
	require.Equal(t, uploadSessionInMemory.FileUploadSessionID, uploadSessionID)
	require.Equal(t, uploadSessionInMemory.TeamID, teamID)
	require.Nil(t, uploadSessionInMemory.UpdatedAt)
	require.Equal(t, uploadSessionInMemory.Type, entity.IconTeamFileUploadSessionType)
	require.Equal(t, uploadSessionInMemory.IsCompleted, false)
}

func TestTeamService_FinishTeamIconUploadSession(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
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
	require.Nil(t, err)

	defer tx.Rollback()

	require.Nil(t, teamRef.teamDao.CreateTeam(ct, tx, team))

	uploadSessionID, internalErr := teamRef.teamService.CreateTeamIconUploadSession(ct, teamID)
	require.Nil(t, internalErr)

	updatedTeam, err := teamRef.teamService.FinishTeamIconUploadSession(ct, team.ID, uploadSessionID)
	require.Nil(t, err)
	require.Equal(t, team.ID, updatedTeam.ID)
	require.NotEqual(t, team.IconURL, updatedTeam.IconURL)
	require.NotEqual(t, team.UpdatedAt, updatedTeam.UpdatedAt)
	require.Equal(t, team.Name, updatedTeam.Name)
	require.Equal(t, team.CreatedAt, updatedTeam.CreatedAt)
	require.Equal(t, team.OwnerUserID, updatedTeam.OwnerUserID)
	require.Equal(t, team.CreatorUserID, updatedTeam.CreatorUserID)

	uploadSessionInMemory, err := teamRef.teamFileUploadSessionDao.FindTeamFileUploadSessionByTeamIDWithTx(ct,
		tx,
		teamID,
		entity.IconTeamFileUploadSessionType,
		uploadSessionID)
	require.Nil(t, err)
	require.Equal(t, uploadSessionInMemory.FileUploadSessionID, uploadSessionID)
	require.Equal(t, uploadSessionInMemory.TeamID, teamID)
	require.NotNil(t, uploadSessionInMemory.UpdatedAt)
	require.Equal(t, uploadSessionInMemory.Type, entity.IconTeamFileUploadSessionType)
	require.Equal(t, uploadSessionInMemory.IsCompleted, true)
}

func TestTeamService_FindTeamMembers(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var teamID1 uint64 = 12
	var teamID2 uint64 = 24
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := teamRef.transactionFactory.BeginTx(ct, nil)
	require.Nil(t, err)

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

	require.Nil(t, teamRef.teamMemberDao.CreateTeamMember(ct, tx, teamMember1))
	require.Nil(t, teamRef.teamMemberDao.CreateTeamMember(ct, tx, teamMember2))

	teamsFound, internalErr := teamRef.teamService.FindTeamMembers(ct, teamMember1.TeamID)
	require.Nil(t, internalErr)
	require.Equal(t, 1, len(teamsFound))
	require.Equal(t, teamMember1.TeamID, teamsFound[0].TeamID)
	require.Equal(t, teamMember1.UserID, teamsFound[0].UserID)
	require.Equal(t, teamMember1.WeeklyBandwidth, teamsFound[0].WeeklyBandwidth)
	require.Equal(t, teamMember1.CreatedAt, teamsFound[0].CreatedAt)
	require.Equal(t, teamMember1.UpdatedAt, teamsFound[0].UpdatedAt)
}

func TestTeamService_AddMemberToTeam(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var teamID uint64 = 12
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := teamRef.transactionFactory.BeginTx(ct, nil)
	require.Nil(t, err)

	defer tx.Rollback()

	teamMember, internalErr := teamRef.teamService.AddMemberToTeam(ct, teamID, requesterUserID)
	require.Nil(t, internalErr)

	require.Equal(t, teamMember.TeamID, teamID)
	require.Equal(t, teamMember.UserID, requesterUserID)

	teamMemberInMemory, internalErr := teamRef.teamMemberDao.FindTeamMemberWithTx(ct, tx, teamID, requesterUserID)
	require.Nil(t, internalErr)
	require.Equal(t, teamMemberInMemory.TeamID, teamID)
	require.Equal(t, teamMemberInMemory.UserID, requesterUserID)
}

func TestTeamService_RemoveMemberFromTeam(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var teamID uint64 = 12
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := teamRef.transactionFactory.BeginTx(ct, nil)
	require.Nil(t, err)

	defer tx.Rollback()

	now := time.Now().UTC()
	teamMember := entity.TeamMember{
		TeamID:          teamID,
		UserID:          requesterUserID,
		WeeklyBandwidth: timePerWeek,
		CreatedAt:       now,
		UpdatedAt:       &now,
	}

	require.Nil(t, teamRef.teamMemberDao.CreateTeamMember(ct, tx, teamMember))

	teamDeleted, internalErr := teamRef.teamService.RemoveMemberFromTeam(ct, teamMember.TeamID, teamMember.UserID)
	require.Nil(t, internalErr)
	require.Equal(t, teamDeleted.TeamID, teamMember.TeamID)
	require.Equal(t, teamDeleted.UserID, teamMember.UserID)
	require.Equal(t, teamDeleted.WeeklyBandwidth, teamMember.WeeklyBandwidth)
	require.Equal(t, teamDeleted.CreatedAt, teamMember.CreatedAt)
	require.Equal(t, teamDeleted.UpdatedAt, teamMember.UpdatedAt)

	_, internalErr = teamRef.teamMemberDao.FindTeamMemberWithTx(ct, tx, teamMember.TeamID, teamMember.UserID)
	require.NotNil(t, internalErr)
	require.Equal(t, internalErr.Code, errs.NotFound)
}

func TestTeamService_UpdateTeamMember(t *testing.T) {
	teamRef, ok := prepareTeamTestRef(t, feature.Toggles{
		EnableAuthorization: false,
	})
	if !ok {
		return
	}

	var requesterUserID uint64 = 20
	var teamID uint64 = 12
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, requesterUserID)

	tx, err := teamRef.transactionFactory.BeginTx(ct, nil)
	require.Nil(t, err)

	defer tx.Rollback()

	now := time.Now().UTC()
	teamMember := entity.TeamMember{
		TeamID:          teamID,
		UserID:          requesterUserID,
		WeeklyBandwidth: timePerWeek,
		CreatedAt:       now,
		UpdatedAt:       &now,
	}

	require.Nil(t, teamRef.teamMemberDao.CreateTeamMember(ct, tx, teamMember))

	updateInput := UpdateTeamMemberInput{
		UserID:          requesterUserID,
		WeeklyBandwidth: timePerWeek / 7,
	}
	teamUpdated, internalErr := teamRef.teamService.UpdateTeamMember(ct, teamMember.TeamID, updateInput)
	require.Nil(t, internalErr)
	require.Equal(t, teamUpdated.TeamID, teamMember.TeamID)
	require.Equal(t, teamUpdated.UserID, teamMember.UserID)
	require.Equal(t, teamUpdated.WeeklyBandwidth, updateInput.WeeklyBandwidth)
	require.Equal(t, teamUpdated.CreatedAt, teamMember.CreatedAt)
	require.NotEqual(t, teamUpdated.UpdatedAt, teamMember.UpdatedAt)

	teamMemberInMemory, internalErr := teamRef.teamMemberDao.FindTeamMemberWithTx(ct, tx, teamMember.TeamID, teamMember.UserID)
	require.Nil(t, internalErr)
	require.Equal(t, teamMemberInMemory.TeamID, teamMember.TeamID)
	require.Equal(t, teamMemberInMemory.UserID, teamMember.UserID)
	require.Equal(t, teamMemberInMemory.WeeklyBandwidth, updateInput.WeeklyBandwidth)
	require.Equal(t, teamMemberInMemory.CreatedAt, teamMember.CreatedAt)
	require.NotEqual(t, teamMemberInMemory.UpdatedAt, teamMember.UpdatedAt)
}

func prepareTeamTestRef(t *testing.T, toggles feature.Toggles) (TeamTestRef, bool) {
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
	apiToken, internalErr := servicetest.GetServiceAccountAPIToken(cloudTestKit.IdentityService)
	require.Nil(t, internalErr)

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
	teamyBackendDB.CreateTable(daotest.TeamTableName)
	teamyBackendDB.CreateTable(daotest.TeamMemberTableName)
	teamyBackendDB.CreateTable(daotest.TeamFileUploadSessionTableName)
	teamyBackendDB.CreateTable(daotest.TeamMemberGroupTableName)
	teamyBackendDB.CreateTable(daotest.TeamMemberGroupUserRelationTableName)
	teamyBackendDB.CreateTable(daotest.SprintTableName)

	teamMemberDao := daotest.NewTeamMember(teamyBackendDB, transactionFactory)
	stateSyncer := realtime.NewStateSyncer(logger, teamMemberDao)

	taskDao := daotest.NewTask(teamyBackendDB, transactionFactory)
	sprintDao := daotest.NewSprint(teamyBackendDB, transactionFactory)
	sprintParticipantDao := daotest.NewSprintParticipant(teamyBackendDB, transactionFactory)
	teamDao := daotest.NewTeam(teamyBackendDB, transactionFactory)
	teamFileUploadSessionDao := daotest.NewTeamFileUploadSession(teamyBackendDB)
	teamMemberGroupDao := daotest.NewTeamMemberGroup(teamyBackendDB, transactionFactory)
	teamMemberGroupUserRelationDao := daotest.NewTeamMemberGroupUserRelation(teamyBackendDB, transactionFactory)
	teamMemberGroupRepo := repository.NewTeamMemberGroup(teamMemberGroupDao, teamMemberGroupUserRelationDao)
	teamService := NewTeam(
		logger,
		cloudTestKitConfig.WebAPIBaseURL,
		cloudClientRegistry,
		authorizer,
		toggles,
		stateSyncer,
		transactionFactory,
		taskDao,
		sprintDao,
		sprintParticipantDao,
		teamDao,
		teamMemberDao,
		teamFileUploadSessionDao,
		teamMemberGroupDao,
		teamMemberGroupUserRelationDao,
		teamMemberGroupRepo,
	)
	return TeamTestRef{
		teamService:                    teamService,
		teamDao:                        teamDao,
		teamMemberDao:                  teamMemberDao,
		teamFileUploadSessionDao:       teamFileUploadSessionDao,
		teamMemberGroupDao:             teamMemberGroupDao,
		teamMemberGroupUserRelationDao: teamMemberGroupUserRelationDao,
		cloudTestKit:                   cloudTestKit,
	}, true
}
