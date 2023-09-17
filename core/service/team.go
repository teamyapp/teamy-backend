package service

import (
	"context"
	"fmt"
	"time"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/client"
	cloudAuthorization "github.com/teamyapp/cloud/libs/authorization"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudTransaction "github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/transaction"
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
	cloudClientRegistry        *client.Registry
	authorizer                 client.Authorizer
	featureToggles             feature.Toggles
	stateSyncer                *realtime.StateSyncer
	transactionFactory         cloudTransaction.Factory
	taskDao                    dao.Task
	sprintDao                  dao.Sprint
	sprintParticipantDao       dao.SprintParticipant
	teamDao                    dao.Team
	teamMemberDao              dao.TeamMember
	teamFileUploadSessionDao   dao.TeamFileUploadSession
	teamGroupDao               dao.TeamGroup
}

func (t Team) FindTeamByID(ct context.Context, teamID uint64) (entity.Team, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Team{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if t.featureToggles.EnableAuthorization {
		query := authorization.NewReadInTeamQuery(userID, teamID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Team{}, err
		}

		if !hasPermission {
			return entity.Team{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	return t.teamDao.FindTeamByID(ct, teamID)
}

func (t Team) FindTeams(ct context.Context, filter *TeamFilter) ([]entity.Team, *errs.Error) {
	teams, err := t.teamDao.FindAllTeams(ct)
	if err != nil {
		return nil, err
	}

	if t.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return nil, nil
		}

		authorizedTeams, err := client.FilterAuthorizedItems(
			ct,
			t.authorizer,
			teams,
			func(team entity.Team) cloudAuthorization.Query {
				return authorization.NewReadInTeamQuery(userID, team.ID)
			})
		if err != nil {
			return nil, err
		}

		teams = authorizedTeams
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
	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		ids, internalErr := t.teamMemberDao.FindTeamIDsByUserIDWithTx(ct, tx, userID)
		if internalErr != nil {
			return internalErr
		}

		if len(ids) < 1 {
			return nil
		}

		teams, internalErr = t.teamDao.FindTeamsByIDsWithTx(ct, tx, ids)
		if internalErr != nil {
			return internalErr
		}

		if t.featureToggles.EnableAuthorization {
			userID, ok := ctx.UserIDFromContext(ct)
			if !ok {
				return nil
			}

			authorizedTeams, err := client.FilterAuthorizedItems(
				ct,
				t.authorizer,
				teams,
				func(team entity.Team) cloudAuthorization.Query {
					return authorization.NewReadInTeamQuery(userID, team.ID)
				})
			if err != nil {
				return err
			}

			teams = authorizedTeams
		}

		if filter != nil {
			teams = filterTeams(teams, *filter)
		}

		return nil
	})

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

	teamID := genTeamIDRes.UniqueNumber
	var ownerGroupID uint64
	var adminGroupID uint64
	var memberGroupID uint64
	var err *errs.Error
	if t.featureToggles.EnableAuthorization {
		err = t.authorizer.RegisterResource(ct, authorization.TeamResourceType, teamID)
		if err != nil {
			return entity.Team{}, err
		}

		teamOwnerUserGroupName := fmt.Sprintf("Team%d/Owner", teamID)
		teamOwnerDescription := fmt.Sprintf("Owners for %s", teamOwnerUserGroupName)
		teamOwnerOperations := make([]cloudAuthorization.ResourceOperation, 0)
		for _, teamOwnerResourceTypeOperation := range authorization.TeamOwnerResourceTypeOperations {
			teamOwnerOperations = append(teamOwnerOperations, cloudAuthorization.ResourceOperation{
				ResourceType: teamOwnerResourceTypeOperation.ResourceType,
				Operation:    teamOwnerResourceTypeOperation.Operation,
				ResourceID:   teamID,
			})
		}
		ownerGroupID, err = t.authorizer.CreateUserGroupAndAssignPermissions(ct,
			userID,
			teamOwnerUserGroupName,
			&teamOwnerDescription,
			teamOwnerOperations,
		)
		if err != nil {
			return entity.Team{}, err
		}

		teamAdminUserGroupName := fmt.Sprintf("Team%d/Admin", teamID)
		teamAdminDescription := fmt.Sprintf("Admins for %s", teamAdminUserGroupName)
		teamAdminOperations := make([]cloudAuthorization.ResourceOperation, 0)
		for _, teamAdminResourceTypeOperation := range authorization.TeamAdminResourceTypeOperations {
			teamAdminOperations = append(teamAdminOperations, cloudAuthorization.ResourceOperation{
				ResourceType: teamAdminResourceTypeOperation.ResourceType,
				Operation:    teamAdminResourceTypeOperation.Operation,
				ResourceID:   teamID,
			})
		}

		adminGroupID, err = t.authorizer.CreateUserGroupAndAssignPermissions(ct,
			userID,
			teamAdminUserGroupName,
			&teamAdminDescription,
			teamAdminOperations,
		)
		if err != nil {
			return entity.Team{}, err
		}

		teamMemberUserGroupName := fmt.Sprintf("Team%d/Member", teamID)
		teamMemberDescription := fmt.Sprintf("Members for %s", teamMemberUserGroupName)
		teamMemberOperations := make([]cloudAuthorization.ResourceOperation, 0)
		for _, teamMemberResourceTypeOperation := range authorization.TeamMemberResourceTypeOperations {
			teamMemberOperations = append(teamMemberOperations, cloudAuthorization.ResourceOperation{
				ResourceType: teamMemberResourceTypeOperation.ResourceType,
				Operation:    teamMemberResourceTypeOperation.Operation,
				ResourceID:   teamID,
			})
		}

		memberGroupID, err = t.authorizer.CreateUserGroupAndAssignPermissions(ct,
			userID,
			teamMemberUserGroupName,
			&teamMemberDescription,
			teamMemberOperations,
		)
		if err != nil {
			return entity.Team{}, err
		}
	}

	team := entity.Team{
		ID:            teamID,
		Name:          input.Name,
		CreatorUserID: userID,
		OwnerUserID:   userID,
		CreatedAt:     time.Now(),
	}
	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err = txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		// All users are authorized to create team
		createTeamMutation := mutation.NewCreateTeam(
			t.logger,
			t.stateSyncer,
			t.teamDao,
			team,
		)
		internalErr := createTeamMutation.Execute(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		rtTx.AppendMutation(createTeamMutation)
		teamMember := entity.TeamMember{
			TeamID:    team.ID,
			UserID:    userID,
			CreatedAt: time.Now(),
		}
		createTeamMemberMutation := mutation.NewCreateTeamMember(
			t.logger,
			t.stateSyncer,
			t.teamMemberDao,
			teamMember,
		)
		internalErr = createTeamMemberMutation.Execute(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		rtTx.AppendMutation(createTeamMemberMutation)

		if t.featureToggles.EnableAuthorization {
			teamGroups := []entity.TeamGroup{
				{
					TeamID:      teamID,
					Label:       entity.OwnerTeamGroupLabel,
					UserGroupID: ownerGroupID,
					CreatedAt:   time.Now().UTC(),
				},
				{
					TeamID:      teamID,
					Label:       entity.AdminTeamGroupLabel,
					UserGroupID: adminGroupID,
					CreatedAt:   time.Now().UTC(),
				},
				{
					TeamID:      teamID,
					Label:       entity.MemberTeamGroupLabel,
					UserGroupID: memberGroupID,
					CreatedAt:   time.Now().UTC(),
				},
			}
			for _, teamGroup := range teamGroups {
				teamGroupMutation := mutation.NewCreateTeamGroup(
					t.logger,
					t.stateSyncer,
					t.teamGroupDao,
					teamGroup,
				)
				internalErr = teamGroupMutation.Execute(ct, tx)
				if internalErr != nil {
					return internalErr
				}

				rtTx.AppendMutation(teamGroupMutation)
			}

			return nil
		}

		return nil
	})

	return team, err
}

func (t Team) UpdateTeam(ct context.Context, teamID uint64, input UpdateTeamInput) (entity.Team, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Team{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if t.featureToggles.EnableAuthorization {
		query := authorization.NewUpdateInTeamQuery(userID, teamID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Team{}, err
		}

		if !hasPermission {
			return entity.Team{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var team entity.Team
	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		team, internalErr = t.teamDao.FindTeamByIDWithTx(ct, tx, teamID)
		if internalErr != nil {
			return internalErr
		}

		team.Name = input.Name
		team.OwnerUserID = input.OwnerUserID
		updatedAt := time.Now().UTC()
		team.UpdatedAt = &updatedAt
		updateTeamMutation := mutation.NewUpdateTeam(
			t.logger,
			t.stateSyncer,
			t.teamDao,
			team,
		)

		internalErr = updateTeamMutation.Execute(ct, tx)
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
	if t.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.Team{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewDeleteInTeamQuery(userID, teamID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Team{}, err
		}

		if !hasPermission {
			return entity.Team{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("authorization query: %v", query))
		}
	}

	var team entity.Team
	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		team, internalErr = t.teamDao.FindTeamByIDWithTx(ct, tx, teamID)
		if internalErr != nil {
			return internalErr
		}

		deleteTeamMutation := mutation.NewDeleteTeam(
			t.logger,
			t.stateSyncer,
			t.teamDao,
			teamID,
		)
		internalErr = deleteTeamMutation.Execute(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		rtTx.AppendMutation(deleteTeamMutation)
		return nil
	})

	if err != nil {
		return entity.Team{}, err
	}

	// TODO: clean up resource relations in authorization service
	return team, nil
}

func (t Team) CreateTeamIconUploadSession(ct context.Context, teamID uint64) (uint64, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return 0, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if t.featureToggles.EnableAuthorization {
		query := authorization.NewUpdateInTeamQuery(userID, teamID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
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
	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		return t.teamFileUploadSessionDao.CreateTeamFileUploadSession(ct, tx, fileUploadSession)
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

	if t.featureToggles.EnableAuthorization {
		query := authorization.NewUpdateInTeamQuery(userID, teamID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
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
	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		iconUploadSession, internalErr = t.teamFileUploadSessionDao.FindTeamFileUploadSessionByTeamIDWithTx(
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
		internalErr = t.teamFileUploadSessionDao.UpdateTeamFileUploadSession(ct, tx, iconUploadSession)
		if internalErr != nil {
			return internalErr
		}

		team, internalErr = t.teamDao.FindTeamByIDWithTx(ct, tx, teamID)
		if internalErr != nil {
			return internalErr
		}

		iconUrl := io.GetFileURL(t.cloudWebAPIExternalBaseURL, uploadSession.FileId)
		team.IconURL = &iconUrl
		team.UpdatedAt = &now
		return t.teamDao.UpdateTeam(ct, tx, team)
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

	if t.featureToggles.EnableAuthorization {
		query := authorization.NewReadMembersInTeamQuery(userID, teamID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
		if err != nil {
			return nil, err
		}

		if !hasPermission {
			return nil, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	return t.teamMemberDao.FindTeamMembersByTeamID(ct, teamID)
}

func (t Team) AddMemberToTeam(ct context.Context, teamID uint64, memberUserID uint64) (entity.TeamMember, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.TeamMember{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if t.featureToggles.EnableAuthorization {
		query := authorization.NewAddMemberToInTeamQuery(userID, teamID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
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
	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		createTeamMemberMutation := mutation.NewCreateTeamMember(
			t.logger,
			t.stateSyncer,
			t.teamMemberDao,
			teamMember,
		)
		internalErr = createTeamMemberMutation.Execute(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		rtTx.AppendMutation(createTeamMemberMutation)

		sprints, internalErr := t.sprintDao.FindSprintsByTeamIDWithTx(ct, tx, teamID)
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
			createSprintParticipantMutation := mutation.NewCreateSprintParticipant(
				t.logger,
				t.stateSyncer,
				t.sprintParticipantDao,
				t.sprintDao,
				participant,
			)
			internalErr = createTeamMemberMutation.Execute(ct, tx)
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

	if t.featureToggles.EnableAuthorization {
		query := authorization.NewRemoveMemberFromInTeamQuery(userID, teamID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.TeamMember{}, err
		}

		if !hasPermission {
			return entity.TeamMember{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var teamMember entity.TeamMember
	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		teamMember, internalErr = t.teamMemberDao.FindTeamMemberWithTx(ct, tx, teamID, memberUserID)
		if internalErr != nil {
			return internalErr
		}

		deleteTeamMemberMutation := mutation.NewDeleteTeamMember(
			t.logger,
			t.stateSyncer,
			t.teamMemberDao,
			teamID,
			teamMember.UserID,
		)
		internalErr = deleteTeamMemberMutation.Execute(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		sprints, internalErr := t.sprintDao.FindSprintsByTeamIDWithTx(ct, tx, teamID)
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
			deleteSprintParticipantMutation := mutation.NewDeleteSprintParticipant(
				t.logger,
				t.stateSyncer,
				t.sprintParticipantDao,
				t.sprintDao,
				teamMember.UserID,
				sprint.ID,
			)
			internalErr = deleteSprintParticipantMutation.Execute(ct, tx)
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

	if t.featureToggles.EnableAuthorization {
		query := authorization.NewUpdateMembersInTeamQuery(userID, teamID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.TeamMember{}, err
		}

		if !hasPermission {
			return entity.TeamMember{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var teamMember entity.TeamMember
	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		teamMember, internalErr = t.teamMemberDao.FindTeamMemberWithTx(ct, tx, teamID, input.UserID)
		if internalErr != nil {
			return internalErr
		}

		bandwidthDelta := input.WeeklyBandwidth - teamMember.WeeklyBandwidth
		teamMember.WeeklyBandwidth = input.WeeklyBandwidth
		now := time.Now().UTC()
		teamMember.UpdatedAt = &now
		updateTeamMemberMutation := mutation.NewUpdateTeamMember(
			t.logger,
			t.stateSyncer,
			t.teamMemberDao,
			teamMember,
		)
		internalErr = updateTeamMemberMutation.Execute(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		sprints, internalErr := t.sprintDao.FindSprintsByTeamIDWithTx(ct, tx, teamID)
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
			participants, err := t.sprintParticipantDao.FindParticipantsBySprintIDWithTx(ct, tx, sprint.ID)
			if err != nil {
				return err
			}

			for _, participant := range participants {
				if participant.UserID != input.UserID {
					continue
				}

				participant.TotalBandwidth += bandwidthDelta
				participant.UnusedBandwidth += bandwidthDelta
				updateSprintParticipantMutation := mutation.NewUpdateSprintParticipant(
					t.logger,
					t.stateSyncer,
					t.sprintParticipantDao,
					t.sprintDao,
					participant,
				)
				internalErr = updateSprintParticipantMutation.Execute(ct, tx)
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
	cloudClientRegistry *client.Registry,
	authorizer client.Authorizer,
	featureToggles feature.Toggles,
	stateSyncer *realtime.StateSyncer,
	transactionFactory cloudTransaction.Factory,
	taskDao dao.Task,
	sprintDao dao.Sprint,
	sprintParticipantDao dao.SprintParticipant,
	teamDao dao.Team,
	teamMemberDao dao.TeamMember,
	teamFileUploadSessionDao dao.TeamFileUploadSession,
	teamGroupDao dao.TeamGroup,
) Team {
	return Team{
		logger:                     logger,
		cloudWebAPIExternalBaseURL: cloudWebAPIExternalBaseURL,
		cloudClientRegistry:        cloudClientRegistry,
		authorizer:                 authorizer,
		featureToggles:             featureToggles,
		stateSyncer:                stateSyncer,
		transactionFactory:         transactionFactory,
		taskDao:                    taskDao,
		sprintDao:                  sprintDao,
		sprintParticipantDao:       sprintParticipantDao,
		teamMemberDao:              teamMemberDao,
		teamDao:                    teamDao,
		teamFileUploadSessionDao:   teamFileUploadSessionDao,
		teamGroupDao:               teamGroupDao,
	}
}
