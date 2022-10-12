package service

import (
	"context"
	"fmt"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/collection"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UpdateTeamMemberInput struct {
	UserID          uint64
	WeeklyBandwidth time.Duration
}

type Team struct {
	dataCollector              obs.DataCollector
	cloudWebAPIExternalBaseURL string
	cloudClientRegistry        *cloudAPI.ClientRegistry
	taskDao                    dao.Task
	sprintDao                  dao.Sprint
	teamDao                    dao.Team
	teamMemberDao              dao.TeamMember
	teamFileUploadSessionDao   dao.TeamFileUploadSession
	teamMemberSyncer           collection.TeamMemberSyncer
	sprintParticipantSyncer    collection.SprintParticipantSyncer
	sprintService              Sprint
}

func (t Team) FindTeamByID(ct context.Context, teamID uint64) (entity.Team, error) {
	return t.teamDao.FindTeamByID(ct, teamID)
}

func (t Team) FindTeams(ct context.Context, filter *TeamFilter) ([]entity.Team, error) {
	teams, err := t.teamDao.FindAllTeams(ct)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	if filter != nil {
		teams = filterTeams(teams, *filter)
	}

	return teams, nil
}

func (t Team) FindTasksInTeam(ct context.Context, teamID uint64, filter *TaskFilter) ([]entity.Task, error) {
	tasks, err := t.taskDao.FindTasksByTeamID(ct, teamID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	if filter != nil {
		tasks = filterTasks(tasks, *filter)
	}

	return tasks, nil
}

func (t Team) FindSprintsInTeam(ct context.Context, teamID uint64, filter *SprintFilter) ([]entity.Sprint, error) {
	sprints, err := t.sprintDao.FindSprintsByTeamID(ct, teamID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	if filter != nil {
		sprints = filterSprints(sprints, *filter)
	}

	return sprints, nil
}

func (t Team) CreateTeamIconUploadSession(ct context.Context, teamID uint64) (uint64, error) {
	res, err := t.cloudClientRegistry.FileClient().CreateUploadSession(ct, &emptypb.Empty{})
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return 0, err
	}

	fileUploadSession := entity.TeamFileUploadSession{
		TeamID:              teamID,
		Type:                entity.IconTeamFileUploadSessionType,
		FileUploadSessionID: res.UploadSessionId,
		IsCompleted:         false,
		CreatedAt:           time.Now(),
	}
	err = t.teamFileUploadSessionDao.CreateTeamFileUploadSession(ct, fileUploadSession)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return 0, err
	}

	return res.UploadSessionId, err
}

func (t Team) FinishTeamIconUploadSession(ct context.Context, teamID uint64, fileUploadSessionID uint64) (entity.Team, error) {
	iconUploadSession, err := t.teamFileUploadSessionDao.FindTeamFileUploadSessionByTeamID(
		ct,
		teamID,
		entity.IconTeamFileUploadSessionType,
		fileUploadSessionID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Team{}, err
	}

	if iconUploadSession.IsCompleted {
		err = fmt.Errorf("icon upload session is already completed: teamID=%v, fileUploadSessionID=%v",
			teamID, fileUploadSessionID)
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Team{}, err
	}

	now := time.Now()
	iconUploadSession.IsCompleted = true
	iconUploadSession.UpdatedAt = &now
	err = t.teamFileUploadSessionDao.UpdateTeamFileUploadSession(ct, iconUploadSession)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Team{}, err
	}

	findUploadSessionReq := proto.FindUploadSessionRequest{
		UploadSessionId: fileUploadSessionID,
	}
	uploadSession, err := t.cloudClientRegistry.FileClient().FindUploadSession(ct, &findUploadSessionReq)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Team{}, err
	}

	team, err := t.teamDao.FindTeamByID(ct, teamID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Team{}, err
	}

	iconUrl := io.GetFileURL(t.cloudWebAPIExternalBaseURL, uploadSession.FileId)
	team.IconURL = &iconUrl
	team.UpdatedAt = &now
	return team, t.teamDao.UpdateTeam(ct, team)
}

func (t Team) AddMemberToTeam(ct context.Context, teamID uint64, memberUserID uint64) (entity.TeamMember, error) {
	teamMember := entity.TeamMember{
		TeamID:    teamID,
		UserID:    memberUserID,
		CreatedAt: time.Now(),
	}
	err := t.teamMemberSyncer.CreateAndSyncTeamMember(ct, teamMember)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.TeamMember{}, err
	}

	currAndFutureSprints, err := t.sprintService.FindCurrentAndFutureSprints(ct, teamID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.TeamMember{}, err
	}

	for _, sprint := range currAndFutureSprints {
		participant := entity.SprintParticipant{
			SprintID:  sprint.ID,
			UserID:    memberUserID,
			CreatedAt: time.Now(),
		}
		err = t.sprintParticipantSyncer.CreateAndSyncSprintParticipant(ct, participant)
		if err != nil {
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.TeamMember{}, err
		}
	}

	return teamMember, nil
}

func (t Team) RemoveMemberFromTeam(ct context.Context, teamID uint64, memberUserID uint64) (entity.TeamMember, error) {
	teamMember, err := t.teamMemberDao.FindTeamMember(ct, teamID, memberUserID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.TeamMember{}, err
	}

	// TODO: ensure user is inside the team
	err = t.teamMemberSyncer.DeleteAndSyncTeamMember(ct, teamID, memberUserID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.TeamMember{}, err
	}

	return teamMember, nil
}

func (t Team) UpdateTeamMember(
	ct context.Context,
	teamID uint64,
	input UpdateTeamMemberInput,
) (entity.TeamMember, error) {
	teamMember, err := t.teamMemberDao.FindTeamMember(ct, teamID, input.UserID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.TeamMember{}, err
	}

	bandwidthDelta := input.WeeklyBandwidth - teamMember.WeeklyBandwidth
	teamMember.WeeklyBandwidth = input.WeeklyBandwidth
	now := time.Now()
	teamMember.UpdatedAt = &now

	err = t.teamMemberDao.UpdateTeamMember(ct, teamMember)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.TeamMember{}, err
	}

	currAndFutureSprints, err := t.sprintService.FindCurrentAndFutureSprints(ct, teamID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.TeamMember{}, err
	}

	for _, sprint := range currAndFutureSprints {
		participants, err := t.sprintService.FindParticipantsInSprint(ct, sprint.ID)
		if err != nil {
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.TeamMember{}, err
		}

		for _, participant := range participants {
			if participant.UserID != input.UserID {
				continue
			}

			participant.TotalBandwidth += bandwidthDelta
			participant.UnusedBandwidth += bandwidthDelta
			err = t.sprintParticipantSyncer.UpdateAndSyncSprintParticipant(ct, participant)
			if err != nil {
				t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
				return entity.TeamMember{}, err
			}
		}
	}

	return teamMember, nil
}

func NewTeam(
	dataCollector obs.DataCollector,
	cloudWebAPIExternalBaseURL string,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	taskDao dao.Task,
	sprintDao dao.Sprint,
	teamDao dao.Team,
	teamMemberDao dao.TeamMember,
	teamFileUploadSessionDao dao.TeamFileUploadSession,
	teamMemberSyncer collection.TeamMemberSyncer,
	sprintParticipantSyncer collection.SprintParticipantSyncer,
	sprintService Sprint,
) Team {
	return Team{
		dataCollector:              dataCollector,
		cloudWebAPIExternalBaseURL: cloudWebAPIExternalBaseURL,
		cloudClientRegistry:        cloudClientRegistry,
		taskDao:                    taskDao,
		sprintDao:                  sprintDao,
		teamDao:                    teamDao,
		teamMemberDao:              teamMemberDao,
		teamFileUploadSessionDao:   teamFileUploadSessionDao,
		teamMemberSyncer:           teamMemberSyncer,
		sprintParticipantSyncer:    sprintParticipantSyncer,
		sprintService:              sprintService,
	}
}
