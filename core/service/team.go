package service

import (
	"context"
	"fmt"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/daov2"
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
	logger                     telemetry.Logger
	cloudWebAPIExternalBaseURL string
	cloudClientRegistry        *cloudAPI.ClientRegistry
	authorizer                 Authorizer
	stateSyncer                *realtime.StateSyncer
	transactionFactory         transaction.Factory
	taskDaoV2                  daov2.Task
	sprintDao                  dao.Sprint
	sprintDaoV2                daov2.Sprint
	sprintParticipantDao       dao.SprintParticipant
	sprintParticipantDaoV2     daov2.SprintParticipant
	teamDao                    dao.Team
	teamDaoV2                  daov2.Team
	teamMemberDao              dao.TeamMember
	teamMemberDaoV2            daov2.TeamMember
	teamFileUploadSessionDaoV2 daov2.TeamFileUploadSession
}

func (t Team) FindTeamByID(ct context.Context, teamID uint64) (entity.Team, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Team{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if feature.EnableAuthorization {
		query := authorization.NewReadTeamQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.Team{}, err
		}

		if !hasPermission {
			return entity.Team{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	return t.teamDaoV2.FindTeamByID(ct, teamID)
}

func (t Team) FindTeams(ct context.Context, filter *TeamFilter) ([]entity.Team, *errs.Error) {
	teams, err := t.teamDaoV2.FindAllTeams(ct)
	if err != nil {
		return nil, err
	}

	if filter != nil {
		teams = filterTeams(teams, *filter)
	}

	return teams, nil
}

func (t Team) FindTeamsForUser(ct context.Context, userID uint64, filter *TeamFilter) ([]entity.Team, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return nil, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	var teams []entity.Team
	txCtx := TransactionsContext{
		logger:             t.logger,
		transactionFactory: t.transactionFactory,
		stateSyncer:        t.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(true, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		ids, internalErr := t.teamMemberDaoV2.FindTeamIDsByUserIDWithTx(ct, tx, userID)
		if internalErr != nil {
			return internalErr
		}

		if len(ids) < 1 {
			return nil
		}

		teams, internalErr = t.teamDaoV2.FindTeamsByIDsWithTx(ct, tx, ids)
		if internalErr != nil {
			return internalErr
		}

		if filter != nil {
			teams = filterTeams(teams, *filter)
		}

		return nil
	})

	// TODO: authorization - need to check permission for each team, and return list of errors
	return teams, err
}

func (t Team) CreateTeam(ct context.Context, input CreateTeamInput) (entity.Team, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Team{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	genTeamIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "teamID"}
	genTeamIDRes, rpcErr := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genTeamIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.Team{}, internalErr
	}

	team := entity.Team{
		ID:            genTeamIDRes.UniqueNumber,
		Name:          input.Name,
		CreatorUserID: userID,
		OwnerUserID:   userID,
		CreatedAt:     time.Now(),
	}
	txCtx := TransactionsContext{
		logger:             t.logger,
		transactionFactory: t.transactionFactory,
		stateSyncer:        t.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		// All users are authorized to create team
		createTeamMutation := mutation.NewCreateTeamMutation(
			t.logger,
			t.stateSyncer,
			t.teamDao,
			t.teamDaoV2,
			team,
		)
		internalErr := createTeamMutation.ExecuteV2(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		rtTx.AppendMutation(createTeamMutation)
		teamMember := entity.TeamMember{
			TeamID:    team.ID,
			UserID:    userID,
			CreatedAt: time.Now(),
		}
		createTeamMemberMutation := mutation.NewCreateTeamMemberMutation(
			t.logger,
			t.stateSyncer,
			t.teamMemberDao,
			t.teamMemberDaoV2,
			teamMember,
		)
		internalErr = createTeamMemberMutation.ExecuteV2(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		rtTx.AppendMutation(createTeamMemberMutation)
		return nil
	})

	if err != nil {
		return entity.Team{}, err
	}

	if feature.EnableAuthorization {
		err = t.authorizer.registerResource(ct, authorization.TeamResourceType, team.ID)
		if err != nil {
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

		_, err = t.authorizer.createUserGroupAndAssignPermissions(ct,
			userID,
			teamAdminUserGroupName,
			&teamAdminDescription,
			teamAdminOperations,
		)
		if err != nil {
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
			return entity.Team{}, err
		}
	}

	return team, nil
}

func (t Team) UpdateTeam(ct context.Context, teamID uint64, input UpdateTeamInput) (entity.Team, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Team{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if feature.EnableAuthorization {
		query := authorization.NewTeamUpdateSettingsQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.Team{}, err
		}

		if !hasPermission {
			return entity.Team{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var team entity.Team
	txCtx := TransactionsContext{
		logger:             t.logger,
		transactionFactory: t.transactionFactory,
		stateSyncer:        t.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		team, internalErr = t.teamDaoV2.FindTeamByIDWithTx(ct, tx, teamID)
		if internalErr != nil {
			return internalErr
		}

		team.Name = input.Name
		team.OwnerUserID = input.OwnerUserID
		updatedAt := time.Now().UTC()
		team.UpdatedAt = &updatedAt
		updateTeamMutation := mutation.NewUpdateTeamMutation(
			t.logger,
			t.stateSyncer,
			t.teamDao,
			t.teamDaoV2,
			team,
		)

		internalErr = updateTeamMutation.ExecuteV2(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		rtTx.AppendMutation(updateTeamMutation)
		return nil
	})

	if err != nil {
		return entity.Team{}, err
	}

	return team, nil
}

func (t Team) DeleteTeam(ct context.Context, teamID uint64) (entity.Team, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.Team{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewDeleteTeamQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.Team{}, err
		}

		if !hasPermission {
			return entity.Team{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("authorization query: %v", query))
		}
	}

	var team entity.Team
	txCtx := TransactionsContext{
		logger:             t.logger,
		transactionFactory: t.transactionFactory,
		stateSyncer:        t.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		team, internalErr = t.teamDaoV2.FindTeamByIDWithTx(ct, tx, teamID)
		if internalErr != nil {
			return internalErr
		}

		deleteTeamMutation := mutation.NewDeleteTeamMutation(
			t.logger,
			t.stateSyncer,
			t.teamDao,
			t.teamDaoV2,
			teamID,
		)
		internalErr = deleteTeamMutation.ExecuteV2(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		rtTx.AppendMutation(deleteTeamMutation)
		return nil
	})

	if err != nil {
		return entity.Team{}, err
	}

	return team, nil
}

func (t Team) CreateTeamIconUploadSession(ct context.Context, teamID uint64) (uint64, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return 0, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if feature.EnableAuthorization {
		query := authorization.NewTeamUpdateSettingsQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			return 0, err
		}

		if !hasPermission {
			return 0, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	res, rpcErr := t.cloudClientRegistry.FileClient().CreateUploadSession(ct, &emptypb.Empty{})
	if rpcErr != nil {
		return 0, errs.FromGRPCErr(rpcErr)
	}

	fileUploadSession := entity.TeamFileUploadSession{
		TeamID:              teamID,
		Type:                entity.IconTeamFileUploadSessionType,
		FileUploadSessionID: res.UploadSessionId,
		IsCompleted:         false,
		CreatedAt:           time.Now(),
	}
	txCtx := TransactionsContext{
		logger:             t.logger,
		transactionFactory: t.transactionFactory,
		stateSyncer:        t.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		return t.teamFileUploadSessionDaoV2.CreateTeamFileUploadSession(ct, tx, fileUploadSession)
	})

	if err != nil {
		return 0, err
	}

	return res.UploadSessionId, err
}

func (t Team) FinishTeamIconUploadSession(ct context.Context, teamID uint64, fileUploadSessionID uint64) (entity.Team, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Team{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if feature.EnableAuthorization {
		query := authorization.NewTeamUpdateSettingsQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.Team{}, err
		}

		if !hasPermission {
			return entity.Team{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	findUploadSessionReq := proto.FindUploadSessionRequest{
		UploadSessionId: fileUploadSessionID,
	}
	uploadSession, rpcErr := t.cloudClientRegistry.FileClient().FindUploadSession(ct, &findUploadSessionReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.Team{}, internalErr
	}

	var iconUploadSession entity.TeamFileUploadSession
	var team entity.Team
	txCtx := TransactionsContext{
		logger:             t.logger,
		transactionFactory: t.transactionFactory,
		stateSyncer:        t.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		iconUploadSession, internalErr = t.teamFileUploadSessionDaoV2.FindTeamFileUploadSessionByTeamIDWithTx(
			ct,
			tx,
			teamID,
			entity.IconTeamFileUploadSessionType,
			fileUploadSessionID)
		if internalErr != nil {
			return internalErr
		}

		if iconUploadSession.IsCompleted {
			return errs.NewError(errs.InvalidOperation, fmt.Sprintf("icon upload session is already completed: teamID=%v, fileUploadSessionID=%v",
				teamID,
				fileUploadSessionID))
		}

		now := time.Now().UTC()
		iconUploadSession.IsCompleted = true
		iconUploadSession.UpdatedAt = &now
		internalErr = t.teamFileUploadSessionDaoV2.UpdateTeamFileUploadSession(ct, tx, iconUploadSession)
		if internalErr != nil {
			return internalErr
		}

		team, internalErr = t.teamDaoV2.FindTeamByIDWithTx(ct, tx, teamID)
		if internalErr != nil {
			return internalErr
		}

		iconUrl := io.GetFileURL(t.cloudWebAPIExternalBaseURL, uploadSession.FileId)
		team.IconURL = &iconUrl
		team.UpdatedAt = &now
		return t.teamDaoV2.UpdateTeam(ct, tx, team)
	})

	if err != nil {
		return entity.Team{}, err
	}

	return team, nil
}

func (t Team) FindTeamMembers(ct context.Context, teamID uint64) ([]entity.TeamMember, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return nil, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if feature.EnableAuthorization {
		query := authorization.NewTeamReadMemberQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			return nil, err
		}

		if !hasPermission {
			return nil, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	return t.teamMemberDaoV2.FindTeamMembersByTeamID(ct, teamID)
}

func (t Team) AddMemberToTeam(ct context.Context, teamID uint64, memberUserID uint64) (entity.TeamMember, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.TeamMember{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if feature.EnableAuthorization {
		query := authorization.NewTeamAddMemberToQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.TeamMember{}, err
		}

		if !hasPermission {
			return entity.TeamMember{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	teamMember := entity.TeamMember{
		TeamID:    teamID,
		UserID:    memberUserID,
		CreatedAt: time.Now(),
	}
	txCtx := TransactionsContext{
		logger:             t.logger,
		transactionFactory: t.transactionFactory,
		stateSyncer:        t.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		createTeamMemberMutation := mutation.NewCreateTeamMemberMutation(
			t.logger,
			t.stateSyncer,
			t.teamMemberDao,
			t.teamMemberDaoV2,
			teamMember,
		)
		internalErr = createTeamMemberMutation.ExecuteV2(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		rtTx.AppendMutation(createTeamMemberMutation)

		sprints, internalErr := t.sprintDaoV2.FindSprintsByTeamIDWithTx(ct, tx, teamID)
		if internalErr != nil {
			return internalErr
		}

		now := time.Now().UTC()
		currAndFutureSprints := collect.Filter(sprints, func(sprint entity.Sprint) bool {
			if sprint.EndAt.UTC().Before(now) {
				return false
			}

			return true
		})

		for _, sprint := range currAndFutureSprints {
			participant := entity.SprintParticipant{
				SprintID:  sprint.ID,
				UserID:    memberUserID,
				CreatedAt: time.Now(),
			}
			createSprintParticipantMutation := mutation.NewCreateSprintParticipantMutation(
				t.logger,
				t.stateSyncer,
				t.sprintParticipantDao,
				t.sprintParticipantDaoV2,
				t.sprintDao,
				t.sprintDaoV2,
				participant,
			)
			internalErr = createTeamMemberMutation.ExecuteV2(ct, tx)
			if internalErr != nil {
				return internalErr
			}

			rtTx.AppendMutation(createSprintParticipantMutation)
		}

		return nil
	})

	if err != nil {
		return entity.TeamMember{}, err
	}

	return teamMember, nil
}

func (t Team) RemoveMemberFromTeam(ct context.Context, teamID uint64, memberUserID uint64) (entity.TeamMember, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.TeamMember{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if feature.EnableAuthorization {
		query := authorization.NewTeamRemoveMemberFromQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.TeamMember{}, err
		}

		if !hasPermission {
			return entity.TeamMember{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var teamMember entity.TeamMember
	txCtx := TransactionsContext{
		logger:             t.logger,
		transactionFactory: t.transactionFactory,
		stateSyncer:        t.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		teamMember, internalErr = t.teamMemberDaoV2.FindTeamMemberWithTx(ct, tx, teamID, memberUserID)
		if internalErr != nil {
			return internalErr
		}

		deleteTeamMemberMutation := mutation.NewDeleteTeamMemberMutation(
			t.logger,
			t.stateSyncer,
			t.teamMemberDao,
			t.teamMemberDaoV2,
			teamID,
			teamMember.UserID,
		)
		internalErr = deleteTeamMemberMutation.ExecuteV2(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		sprints, internalErr := t.sprintDaoV2.FindSprintsByTeamIDWithTx(ct, tx, teamID)
		if internalErr != nil {
			return internalErr
		}

		now := time.Now().UTC()
		currAndFutureSprints := collect.Filter(sprints, func(sprint entity.Sprint) bool {
			if sprint.EndAt.UTC().Before(now) {
				return false
			}

			return true
		})

		for _, sprint := range currAndFutureSprints {
			deleteSprintParticipantMutation := mutation.NewDeleteSprintParticipantMutation(
				t.logger,
				t.stateSyncer,
				t.sprintParticipantDao,
				t.sprintParticipantDaoV2,
				t.sprintDao,
				t.sprintDaoV2,
				teamMember.UserID,
				sprint.ID,
			)
			internalErr = deleteSprintParticipantMutation.ExecuteV2(ct, tx)
			if internalErr != nil {
				return internalErr
			}

			rtTx.AppendMutation(deleteSprintParticipantMutation)
		}

		return nil
	})

	if err != nil {
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
		return entity.TeamMember{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if feature.EnableAuthorization {
		query := authorization.NewTeamUpdateMemberQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.TeamMember{}, err
		}

		if !hasPermission {
			return entity.TeamMember{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var teamMember entity.TeamMember
	txCtx := TransactionsContext{
		logger:             t.logger,
		transactionFactory: t.transactionFactory,
		stateSyncer:        t.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		teamMember, internalErr = t.teamMemberDaoV2.FindTeamMemberWithTx(ct, tx, teamID, input.UserID)
		if internalErr != nil {
			return internalErr
		}

		bandwidthDelta := input.WeeklyBandwidth - teamMember.WeeklyBandwidth
		teamMember.WeeklyBandwidth = input.WeeklyBandwidth
		now := time.Now().UTC()
		teamMember.UpdatedAt = &now
		updateTeamMemberMutation := mutation.NewUpdateTeamMemberMutation(
			t.logger,
			t.stateSyncer,
			t.teamMemberDao,
			t.teamMemberDaoV2,
			teamMember,
		)
		internalErr = updateTeamMemberMutation.ExecuteV2(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		sprints, internalErr := t.sprintDaoV2.FindSprintsByTeamIDWithTx(ct, tx, teamID)
		if internalErr != nil {
			return internalErr
		}

		now = time.Now().UTC()
		currAndFutureSprints := collect.Filter(sprints, func(sprint entity.Sprint) bool {
			if sprint.EndAt.UTC().Before(now) {
				return false
			}

			return true
		})

		for _, sprint := range currAndFutureSprints {
			participants, err := t.sprintParticipantDaoV2.FindParticipantsBySprintIDWithTx(ct, tx, sprint.ID)
			if err != nil {
				return err
			}

			for _, participant := range participants {
				if participant.UserID != input.UserID {
					continue
				}

				participant.TotalBandwidth += bandwidthDelta
				participant.UnusedBandwidth += bandwidthDelta
				updateSprintParticipantMutation := mutation.NewUpdateSprintParticipantMutation(
					t.logger,
					t.stateSyncer,
					t.sprintParticipantDao,
					t.sprintParticipantDaoV2,
					t.sprintDao,
					t.sprintDaoV2,
					participant,
				)
				internalErr = updateSprintParticipantMutation.ExecuteV2(ct, tx)
				if internalErr != nil {
					return internalErr
				}

				rtTx.AppendMutation(updateSprintParticipantMutation)
			}
		}

		return nil
	})

	if err != nil {
		return entity.TeamMember{}, err
	}

	return teamMember, nil
}

func NewTeam(
	logger telemetry.Logger,
	cloudWebAPIExternalBaseURL string,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	authorizer Authorizer,
	stateSyncer *realtime.StateSyncer,
	transactionFactory transaction.Factory,
	taskDaoV2 daov2.Task,
	sprintDao dao.Sprint,
	sprintDaoV2 daov2.Sprint,
	sprintParticipantDao dao.SprintParticipant,
	sprintParticipantDaoV2 daov2.SprintParticipant,
	teamDao dao.Team,
	teamDaoV2 daov2.Team,
	teamMemberDao dao.TeamMember,
	teamMemberDaoV2 daov2.TeamMember,
	teamFileUploadSessionDaoV2 daov2.TeamFileUploadSession,
) Team {
	return Team{
		logger:                     logger,
		cloudWebAPIExternalBaseURL: cloudWebAPIExternalBaseURL,
		cloudClientRegistry:        cloudClientRegistry,
		authorizer:                 authorizer,
		stateSyncer:                stateSyncer,
		transactionFactory:         transactionFactory,
		taskDaoV2:                  taskDaoV2,
		sprintDao:                  sprintDao,
		sprintDaoV2:                sprintDaoV2,
		sprintParticipantDao:       sprintParticipantDao,
		sprintParticipantDaoV2:     sprintParticipantDaoV2,
		teamDao:                    teamDao,
		teamMemberDaoV2:            teamMemberDaoV2,
		teamMemberDao:              teamMemberDao,
		teamDaoV2:                  teamDaoV2,
		teamFileUploadSessionDaoV2: teamFileUploadSessionDaoV2,
	}
}
