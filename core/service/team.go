package service

import (
	"context"
	"fmt"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/cloud/libs/telemetry"
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
	dataCollector              telemetry.DataCollector
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

func (t Team) FindTeamByID(ct context.Context, teamID uint64) (entity.Team, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Team{}, internalErr
	}

	if feature.EnableAuthorization {
		query := authorization.NewReadTeamQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.Team{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("permission denied: authorization query=%v", query),
			}
			t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.Team{}, internalErr
		}
	}

	return t.teamDao.FindTeamByID(ct, teamID)
}

func (t Team) FindTeams(ct context.Context, filter *TeamFilter) ([]entity.Team, *errs.Error) {
	teams, err := t.teamDao.FindAllTeams(ct)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, err
	}

	if filter != nil {
		teams = filterTeams(teams, *filter)
	}

	return teams, nil
}

func (t Team) FindTeamsForUser(ct context.Context, userID uint64, filter *TeamFilter) ([]entity.Team, *errs.Error) {
	ids, err := t.teamMemberDao.FindTeamIDsByUserID(ct, userID)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, err
	}

	if len(ids) < 1 {
		return []entity.Team{}, nil
	}

	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, err
	}

	teams, err := t.teamDao.FindTeamsByIDs(ct, ids)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, err
	}

	if filter != nil {
		teams = filterTeams(teams, *filter)
	}

	// TODO: authorization - need to check permission for each team, and return list of errors
	return teams, nil
}

func (t Team) CreateTeam(ct context.Context, input CreateTeamInput) (entity.Team, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Team{}, internalErr
	}

	genTeamIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "teamID"}
	genTeamIDRes, rpcErr := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genTeamIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Team{}, internalErr
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
	err := realTimeTransaction.ApplyMutation(ct, createTeamMutation)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
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
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Team{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Team{}, err
	}

	if feature.EnableAuthorization {
		err = t.authorizer.registerResource(ct, authorization.TeamResourceType, team.ID)
		if err != nil {
			t.dataCollector.Logger.ErrorWithContext(ct, err)
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
			t.dataCollector.Logger.ErrorWithContext(ct, err)
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
			t.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.Team{}, err
		}
	}

	return team, nil
}

func (t Team) UpdateTeam(ct context.Context, teamID uint64, input UpdateTeamInput) (entity.Team, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Team{}, internalErr
	}

	if feature.EnableAuthorization {
		query := authorization.NewTeamUpdateSettingsQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.Team{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("permission denied: authorization query=%v", query),
			}
			t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.Team{}, internalErr
		}
	}

	team, err := t.teamDao.FindTeamByID(ct, teamID)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
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
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Team{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Team{}, err
	}

	return team, nil
}

func (t Team) DeleteTeam(ct context.Context, teamID uint64) (entity.Team, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			internalErr := &errs.Error{
				Code:    errs.Unauthenticated,
				Message: "user ID not found",
			}

			t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.Team{}, internalErr
		}

		query := authorization.NewDeleteTeamQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			t.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.Team{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("authorization query: %v", query),
			}
			t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.Team{}, internalErr
		}
	}

	team, err := t.teamDao.FindTeamByID(ct, teamID)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Team{}, err
	}

	deleteTeamMutation := mutation.NewDeleteTeamMutation(
		t.dataCollector,
		t.stateSyncer,
		t.teamDao,
		teamID,
	)
	realTimeTransaction := realtime.NewTransaction(t.dataCollector, t.stateSyncer)
	err = realTimeTransaction.ApplyMutation(ct, deleteTeamMutation)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Team{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Team{}, err
	}

	return team, nil
}

func (t Team) CreateTeamIconUploadSession(ct context.Context, teamID uint64) (uint64, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return 0, internalErr
	}

	if feature.EnableAuthorization {
		query := authorization.NewTeamUpdateSettingsQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			return 0, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("permission denied: authorization query=%v", query),
			}
			t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return 0, internalErr
		}
	}

	res, rpcErr := t.cloudClientRegistry.FileClient().CreateUploadSession(ct, &emptypb.Empty{})
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return 0, internalErr
	}

	fileUploadSession := entity.TeamFileUploadSession{
		TeamID:              teamID,
		Type:                entity.IconTeamFileUploadSessionType,
		FileUploadSessionID: res.UploadSessionId,
		IsCompleted:         false,
		CreatedAt:           time.Now(),
	}
	err := t.teamFileUploadSessionDao.CreateTeamFileUploadSession(ct, fileUploadSession)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return 0, err
	}

	return res.UploadSessionId, err
}

func (t Team) FinishTeamIconUploadSession(ct context.Context, teamID uint64, fileUploadSessionID uint64) (entity.Team, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Team{}, internalErr
	}

	if feature.EnableAuthorization {
		query := authorization.NewTeamUpdateSettingsQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.Team{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("permission denied: authorization query=%v", query),
			}
			t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.Team{}, internalErr
		}
	}

	iconUploadSession, err := t.teamFileUploadSessionDao.FindTeamFileUploadSessionByTeamID(
		ct,
		teamID,
		entity.IconTeamFileUploadSessionType,
		fileUploadSessionID)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Team{}, err
	}

	if iconUploadSession.IsCompleted {
		internalErr := &errs.Error{
			Code: errs.InvalidOperation,
			Message: fmt.Sprintf("icon upload session is already completed: teamID=%v, fileUploadSessionID=%v",
				teamID,
				fileUploadSessionID),
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Team{}, internalErr
	}

	now := time.Now()
	iconUploadSession.IsCompleted = true
	iconUploadSession.UpdatedAt = &now
	err = t.teamFileUploadSessionDao.UpdateTeamFileUploadSession(ct, iconUploadSession)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Team{}, err
	}

	findUploadSessionReq := proto.FindUploadSessionRequest{
		UploadSessionId: fileUploadSessionID,
	}
	uploadSession, rpcErr := t.cloudClientRegistry.FileClient().FindUploadSession(ct, &findUploadSessionReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Team{}, internalErr
	}

	team, err := t.teamDao.FindTeamByID(ct, teamID)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Team{}, err
	}

	iconUrl := io.GetFileURL(t.cloudWebAPIExternalBaseURL, uploadSession.FileId)
	team.IconURL = &iconUrl
	team.UpdatedAt = &now
	return team, t.teamDao.UpdateTeam(ct, team)
}

func (t Team) FindTeamMembers(ct context.Context, teamID uint64) ([]entity.TeamMember, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	if feature.EnableAuthorization {
		query := authorization.NewTeamReadMemberQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			t.dataCollector.Logger.ErrorWithContext(ct, err)
			return nil, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("permission denied: authorization query=%v", query),
			}
			t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return nil, internalErr
		}
	}

	return t.teamMemberDao.FindTeamMembersByTeamID(ct, teamID)
}

func (t Team) AddMemberToTeam(ct context.Context, teamID uint64, memberUserID uint64) (entity.TeamMember, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.TeamMember{}, internalErr
	}

	if feature.EnableAuthorization {
		query := authorization.NewTeamAddMemberToQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.TeamMember{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("permission denied: authorization query=%v", query),
			}
			t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.TeamMember{}, internalErr
		}
	}

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
		t.dataCollector.Logger.ErrorWithContext(ct, err)
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
			t.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.TeamMember{}, err
		}
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.TeamMember{}, err
	}

	return teamMember, nil
}

func (t Team) RemoveMemberFromTeam(ct context.Context, teamID uint64, memberUserID uint64) (entity.TeamMember, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.TeamMember{}, internalErr
	}

	if feature.EnableAuthorization {
		query := authorization.NewTeamRemoveMemberFromQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.TeamMember{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("permission denied: authorization query=%v", query),
			}
			t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.TeamMember{}, internalErr
		}
	}

	teamMember, err := t.teamMemberDao.FindTeamMember(ct, teamID, memberUserID)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
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
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.TeamMember{}, err
	}

	currAndFutureSprints, err := t.sprintService.FindCurrentAndFutureSprints(ct, teamID)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
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
			t.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.TeamMember{}, err
		}
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.TeamMember{}, err
	}

	return teamMember, nil
}

func (t Team) UpdateTeamMember(
	ct context.Context,
	teamID uint64,
	input UpdateTeamMemberInput,
) (entity.TeamMember, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.TeamMember{}, internalErr
	}

	if feature.EnableAuthorization {
		query := authorization.NewTeamUpdateMemberQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.TeamMember{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("permission denied: authorization query=%v", query),
			}
			t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.TeamMember{}, internalErr
		}
	}

	teamMember, err := t.teamMemberDao.FindTeamMember(ct, teamID, input.UserID)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
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
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.TeamMember{}, err
	}

	currAndFutureSprints, err := t.sprintService.FindCurrentAndFutureSprints(ct, teamID)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.TeamMember{}, err
	}

	for _, sprint := range currAndFutureSprints {
		participants, err := t.sprintService.FindParticipantsInSprint(ct, sprint.ID)
		if err != nil {
			t.dataCollector.Logger.ErrorWithContext(ct, err)
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
				t.dataCollector.Logger.ErrorWithContext(ct, err)
				return entity.TeamMember{}, err
			}
		}
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		t.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.TeamMember{}, err
	}

	return teamMember, nil
}

func NewTeam(
	dataCollector telemetry.DataCollector,
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
