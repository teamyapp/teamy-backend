package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UpdateTeamMemberInput struct {
	UserID          uint64
	WeeklyBandwidth time.Duration
}

type UpdateTeamInput struct {
	Name        string
	OwnerUserID uint64
}

type CreateTeamInput struct {
	Name string
}

type Team struct {
	dataCollector              obs.DataCollector
	cloudWebAPIExternalBaseURL string
	cloudClientRegistry        *cloudAPI.ClientRegistry
	authorizer                 Authorizer
	stateSyncer                *realtime.StateSyncer
	taskDao                    dao.Task
	sprintDao                  dao.Sprint
	teamDao                    dao.Team
	teamMemberDao              dao.TeamMember
	teamFileUploadSessionDao   dao.TeamFileUploadSession
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

func (t Team) CreateTeam(ct context.Context, input CreateTeamInput) (entity.Team, error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Team{}, err
	}

	genTeamIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "teamID"}
	genTeamIDRes, err := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genTeamIDReq)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Team{}, err
	}

	team := entity.Team{
		ID:            genTeamIDRes.UniqueNumber,
		Name:          input.Name,
		CreatorUserID: userID,
		OwnerUserID:   userID,
		CreatedAt:     time.Now(),
	}

	realTimeTransaction := realtime.NewTransaction(t.dataCollector, t.stateSyncer)
	// All users are authorized to create team
	createTeamMutation := mutation.NewCreateTeamMutation(
		t.dataCollector,
		t.stateSyncer,
		t.teamDao,
		team,
	)
	err = realTimeTransaction.ApplyMutation(ct, createTeamMutation)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Team{}, err
	}

	teamMember := entity.TeamMember{
		TeamID:    team.ID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	createTeamMemberMutation := mutation.NewCreateTeamMemberMutation(
		t.dataCollector,
		t.stateSyncer,
		t.teamMemberDao,
		teamMember,
	)
	err = realTimeTransaction.ApplyMutation(ct, createTeamMemberMutation)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Team{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Team{}, err
	}

	if feature.EnableAuthorization {
		err = t.authorizer.registerResource(ct, authorization.TeamResourceType, team.ID)
		if err != nil {
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Team{}, err
		}

		// When create a new team,
		// 1) Resource: register team resource
		// 2) UserGroup: create TeamAdmin and TeamMember userGroups
		// 3) UserGroupMember:
		// 		add team owner to TeamAdmin group
		// 		add team owner to TeamMember group
		// 4) Permissions:
		//		assign TeamAdmin permissions to team owner
		// 		assign TeamMember permissions to team owner
		teamAdminUserGroupName := fmt.Sprintf("Team%d/Admin", team.ID)
		teamAdminDescription := fmt.Sprintf("Admins for %s", teamAdminUserGroupName)
		teamAdminOperations := make([]authorization.ResourceOperation, 0)
		for _, teamAdminResourceTypeOperation := range authorization.TeamAdminResourceTypeOperations {
			teamAdminOperations = append(teamAdminOperations, authorization.ResourceOperation{
				ResourceType: teamAdminResourceTypeOperation.ResourceType,
				Operation:    teamAdminResourceTypeOperation.Operation,
				ResourceID:   team.ID,
			})
		}

		_, err := t.authorizer.createUserGroupAndAssignPermissions(ct,
			userID,
			teamAdminUserGroupName,
			&teamAdminDescription,
			teamAdminOperations,
		)
		if err != nil {
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Team{}, err
		}

		teamMemberUserGroupName := fmt.Sprintf("Team%d/Member", team.ID)
		teamMemberDescription := fmt.Sprintf("Members for %s", teamMemberUserGroupName)
		teamMemberOperations := make([]authorization.ResourceOperation, 0)
		for _, teamMemberResourceTypeOperation := range authorization.TeamMemberResourceTypeOperations {
			teamMemberOperations = append(teamMemberOperations, authorization.ResourceOperation{
				ResourceType: teamMemberResourceTypeOperation.ResourceType,
				Operation:    teamMemberResourceTypeOperation.Operation,
				ResourceID:   team.ID,
			})
		}

		_, err = t.authorizer.createUserGroupAndAssignPermissions(ct,
			userID,
			teamMemberUserGroupName,
			&teamMemberDescription,
			teamMemberOperations,
		)
		if err != nil {
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Team{}, err
		}
	}

	return team, nil
}

func (t Team) UpdateTeam(ct context.Context, teamID uint64, input UpdateTeamInput) (entity.Team, error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Team{}, err
	}

	if feature.EnableAuthorization {
		query := authorization.NewUpdateTeamSettingsQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.Team{}, err
		}

		if !hasPermission {
			return entity.Team{}, authorization.Error{
				Code:    authorization.UnauthorizedErrorCode,
				Message: fmt.Sprintf("Unauthorized: %v", query),
			}
		}
	}

	team, err := t.teamDao.FindTeamByID(ct, teamID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Team{}, err
	}

	team.Name = input.Name
	team.OwnerUserID = input.OwnerUserID
	updatedAt := time.Now()
	team.UpdatedAt = &updatedAt
	realTimeTransaction := realtime.NewTransaction(t.dataCollector, t.stateSyncer)
	updateTeamMutation := mutation.NewUpdateTeamMutation(
		t.dataCollector,
		t.stateSyncer,
		t.teamDao,
		team,
	)
	err = realTimeTransaction.ApplyMutation(ct, updateTeamMutation)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Team{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Team{}, err
	}

	return team, nil
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
	realTimeTransaction := realtime.NewTransaction(t.dataCollector, t.stateSyncer)
	createTeamMemberMutation := mutation.NewCreateTeamMemberMutation(
		t.dataCollector,
		t.stateSyncer,
		t.teamMemberDao,
		teamMember,
	)
	realTimeTransaction.ApplyMutation(ct, createTeamMemberMutation)

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
		createSprintParticipantMutation := mutation.NewCreateSprintParticipantMutation(
			t.dataCollector,
			t.stateSyncer,
			t.sprintService.sprintParticipantDao,
			t.sprintDao,
			participant,
		)
		err = realTimeTransaction.ApplyMutation(ct, createSprintParticipantMutation)
		if err != nil {
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.TeamMember{}, err
		}
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.TeamMember{}, err
	}

	return teamMember, nil
}

func (t Team) RemoveMemberFromTeam(ct context.Context, teamID uint64, memberUserID uint64) (entity.TeamMember, error) {
	teamMember, err := t.teamMemberDao.FindTeamMember(ct, teamID, memberUserID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.TeamMember{}, err
	}
	realTimeTransaction := realtime.NewTransaction(t.dataCollector, t.stateSyncer)
	// TODO: ensure user is inside the team
	deleteTeamMemberMutation := mutation.NewDeleteTeamMemberMutation(
		t.dataCollector,
		t.stateSyncer,
		t.teamMemberDao,
		teamID,
		teamMember.UserID,
	)
	err = realTimeTransaction.ApplyMutation(ct, deleteTeamMemberMutation)
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
		deleteSprintParticipantMutation := mutation.NewDeleteSprintParticipantMutation(
			t.dataCollector,
			t.stateSyncer,
			t.sprintService.sprintParticipantDao,
			t.sprintDao,
			teamMember.UserID,
			sprint.ID,
		)
		err = realTimeTransaction.ApplyMutation(ct, deleteSprintParticipantMutation)
		if err != nil {
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.TeamMember{}, err
		}
	}

	err = realTimeTransaction.Commit(ct)
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
	realTimeTransaction := realtime.NewTransaction(t.dataCollector, t.stateSyncer)
	updateTeamMemberMutation := mutation.NewUpdateTeamMemberMutation(
		t.dataCollector,
		t.stateSyncer,
		t.teamMemberDao,
		teamMember,
	)
	err = realTimeTransaction.ApplyMutation(ct, updateTeamMemberMutation)
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
			updateSprintParticipantMutation := mutation.NewUpdateSprintParticipantMutation(
				t.dataCollector,
				t.stateSyncer,
				t.sprintService.sprintParticipantDao,
				t.sprintDao,
				participant,
			)
			err = realTimeTransaction.ApplyMutation(ct, updateSprintParticipantMutation)
			if err != nil {
				t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
				return entity.TeamMember{}, err
			}
		}
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.TeamMember{}, err
	}

	return teamMember, nil
}

func NewTeam(
	dataCollector obs.DataCollector,
	cloudWebAPIExternalBaseURL string,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	authorizer Authorizer,
	stateSyncer *realtime.StateSyncer,
	taskDao dao.Task,
	sprintDao dao.Sprint,
	teamDao dao.Team,
	teamMemberDao dao.TeamMember,
	teamFileUploadSessionDao dao.TeamFileUploadSession,
	sprintService Sprint,
) Team {
	return Team{
		dataCollector:              dataCollector,
		cloudWebAPIExternalBaseURL: cloudWebAPIExternalBaseURL,
		cloudClientRegistry:        cloudClientRegistry,
		authorizer:                 authorizer,
		stateSyncer:                stateSyncer,
		taskDao:                    taskDao,
		sprintDao:                  sprintDao,
		teamDao:                    teamDao,
		teamMemberDao:              teamMemberDao,
		teamFileUploadSessionDao:   teamFileUploadSessionDao,
		sprintService:              sprintService,
	}
}
