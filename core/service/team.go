package service

import (
	"context"
	"errors"
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
	"github.com/teamyapp/teamy-backend/core/cache"
	"github.com/teamyapp/teamy-backend/core/dao"
	daoEntity "github.com/teamyapp/teamy-backend/core/dao/entity"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/repository"
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

type CreateTeamMemberGroupInput struct {
	Name   string
	TeamID uint64
}

type UpdateTeamMemberGroupInput struct {
	GroupID uint64
	Name    string
}

type Team struct {
	logger                         telemetry.Logger
	transactionGroupFactory        transaction.GroupFactory
	cloudWebAPIExternalBaseURL     string
	cloudClientRegistry            *client.Registry
	authorizer                     client.Authorizer
	featureToggles                 feature.Toggles
	stateSyncer                    *realtime.StateSyncer
	transactionFactory             cloudTransaction.Factory
	cache                          *cache.TimeBasedCache[string, any]
	taskDao                        dao.Task
	sprintDao                      dao.Sprint
	sprintParticipantDao           dao.SprintParticipant
	teamDao                        dao.Team
	userDao                        dao.User
	teamMemberDao                  dao.TeamMember
	teamFileUploadSessionDao       dao.TeamFileUploadSession
	teamMemberGroupDao             dao.TeamMemberGroup
	teamMemberGroupUserRelationDao dao.TeamMemberGroupUserRelation
	teamMemberGroupRepo            repository.TeamMemberGroup
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

	if t.featureToggles.EnableCache {
		value, cacheErr := t.cache.Get(ct, findTeamCacheKey(teamID))
		if cacheErr == nil {
			return value.(entity.Team), nil
		}

		var cacheKeyNotFoundErr cache.KeyNotFoundErr[string]
		if !errors.As(cacheErr, &cacheKeyNotFoundErr) {
			return entity.Team{}, errs.NewError(errs.Unknown, cacheErr.Error())
		}
	}

	team, err := t.teamDao.FindTeamByID(ct, teamID)
	if err != nil {
		return entity.Team{}, err
	}

	if t.featureToggles.EnableCache {
		cacheErr := t.cache.SetIfExpired(ct, findTeamCacheKey(teamID), team)
		if cacheErr != nil {
			return entity.Team{}, errs.NewError(errs.Unknown, cacheErr.Error())
		}
	}

	return team, nil
}

func (t Team) FindTeams(ct context.Context, filter *TeamFilter) ([]entity.Team, *errs.Error) {
	if t.featureToggles.EnableCache {
		value, cacheErr := t.cache.Get(ct, findTeamsCacheKey(filter))
		if cacheErr == nil {
			return value.([]entity.Team), nil
		}

		var cacheKeyNotFoundErr cache.KeyNotFoundErr[string]
		if !errors.As(cacheErr, &cacheKeyNotFoundErr) {
			return nil, errs.NewError(errs.Unknown, cacheErr.Error())
		}
	}

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

	if t.featureToggles.EnableCache {
		cacheErr := t.cache.SetIfExpired(ct, findTeamsCacheKey(filter), teams)
		if cacheErr != nil {
			return nil, errs.NewError(errs.Unknown, cacheErr.Error())
		}
	}

	return teams, nil
}

func (t Team) FindTeamsForUser(ct context.Context, userID uint64, filter *TeamFilter) ([]entity.Team, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return nil, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	var teams []entity.Team
	err := t.transactionGroupFactory.WithTransactionGroup(
		ct,
		true,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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

	var teamMemberGroups []daoEntity.TeamMemberGroup
	if t.featureToggles.EnableAuthorization {
		genUniqueNumReq := &proto.GenerateUniqueNumberRequest{SequenceName: "teamMemberGroupID"}
		ownerTeamMemberGroupIDRes, err := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genUniqueNumReq)
		if err != nil {
			return entity.Team{}, errs.FromGRPCErr(err)
		}

		adminTeamMemberGroupIDRes, err := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genUniqueNumReq)
		if err != nil {
			return entity.Team{}, errs.FromGRPCErr(err)
		}

		memberTeamMemberGroupIDRes, err := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genUniqueNumReq)
		if err != nil {
			return entity.Team{}, errs.FromGRPCErr(err)
		}

		teamMemberGroups = []daoEntity.TeamMemberGroup{
			{
				ID:                       ownerTeamMemberGroupIDRes.UniqueNumber,
				TeamID:                   teamID,
				Name:                     "Owner",
				AuthorizationUserGroupID: ownerGroupID,
				CreatedAt:                time.Now().UTC(),
			},
			{
				ID:                       adminTeamMemberGroupIDRes.UniqueNumber,
				TeamID:                   teamID,
				Name:                     "Admin",
				AuthorizationUserGroupID: adminGroupID,
				CreatedAt:                time.Now().UTC(),
			},
			{
				ID:                       memberTeamMemberGroupIDRes.UniqueNumber,
				TeamID:                   teamID,
				Name:                     "Member",
				AuthorizationUserGroupID: memberGroupID,
				CreatedAt:                time.Now().UTC(),
			},
		}
	}
	err = t.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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
				for _, teamMemberGroup := range teamMemberGroups {
					teamGroupMutation := mutation.NewCreateTeamGroup(
						t.logger,
						t.stateSyncer,
						t.teamMemberGroupDao,
						teamMemberGroup,
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
	err := t.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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
	err := t.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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
	err := t.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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
	err := t.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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
	err := t.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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
	err := t.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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
	err := t.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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

func (t Team) FindTeamMemberGroups(ct context.Context, teamID uint64) ([]entity.TeamMemberGroup, *errs.Error) {
	var teamMemberGroups []entity.TeamMemberGroup
	err := t.transactionGroupFactory.WithTransactionGroup(
		ct,
		true,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var internalErr *errs.Error
			teamMemberGroups, internalErr = t.teamMemberGroupRepo.FindMemberGroupsByTeamID(ct, tx, teamID)
			return internalErr
		})
	return teamMemberGroups, err
}

func (t Team) FindTeamMemberGroupByID(ct context.Context, id uint64) (entity.TeamMemberGroup, *errs.Error) {
	var teamMemberGroup entity.TeamMemberGroup
	err := t.transactionGroupFactory.WithTransactionGroup(
		ct,
		true,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var internalErr *errs.Error
			teamMemberGroup, internalErr = t.teamMemberGroupRepo.FindMemberGroupByID(ct, tx, id)
			return internalErr
		})
	return teamMemberGroup, err
}

func (t Team) CreateTeamMemberGroup(ct context.Context, input CreateTeamMemberGroupInput) (entity.TeamMemberGroup, *errs.Error) {
	genTeamMemberGroupIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "teamMemberGroupID"}
	genTeamMemberGroupIDRes, rpcErr := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genTeamMemberGroupIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.TeamMemberGroup{}, internalErr
	}

	createUserGroupRes, rpcErr := t.cloudClientRegistry.AuthorizationClient().CreateUserGroup(ct, &proto.CreateUserGroupRequest{
		Name: fmt.Sprintf("Team(%d)/%s", input.TeamID, input.Name),
	})
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.TeamMemberGroup{}, internalErr
	}

	teamMemberGroupID := genTeamMemberGroupIDRes.UniqueNumber
	teamMemberGroupPartial := daoEntity.TeamMemberGroup{
		ID:                       teamMemberGroupID,
		Name:                     input.Name,
		TeamID:                   input.TeamID,
		AuthorizationUserGroupID: createUserGroupRes.UserGroup.GroupId,
		CreatedAt:                time.Now().UTC(),
	}
	internalErr := t.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			return t.teamMemberGroupDao.CreateMemberGroup(ct, tx, teamMemberGroupPartial)
		})
	return repository.GetTeamMemberGroupFromRawTeamMemberGroup(teamMemberGroupPartial), internalErr
}

func (t Team) UpdateTeamMemberGroup(ct context.Context, input UpdateTeamMemberGroupInput) (entity.TeamMemberGroup, *errs.Error) {
	var rawTeamMemberGroup daoEntity.TeamMemberGroup
	internalErr := t.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var internalErr *errs.Error
			rawTeamMemberGroup, internalErr = t.teamMemberGroupDao.FindMemberGroupByID(ct, tx, input.GroupID)
			if internalErr != nil {
				return internalErr
			}

			rawTeamMemberGroup.Name = input.Name
			rawTeamMemberGroup.ID = input.GroupID
			now := time.Now().UTC()
			rawTeamMemberGroup.UpdatedAt = &now
			return t.teamMemberGroupDao.UpdateMemberGroup(ct, tx, rawTeamMemberGroup)
		})
	return repository.GetTeamMemberGroupFromRawTeamMemberGroup(rawTeamMemberGroup), internalErr
}

func (t Team) DeleteTeamMemberGroup(ct context.Context, id uint64) (entity.TeamMemberGroup, *errs.Error) {
	var teamMemberGroup entity.TeamMemberGroup
	internalErr := t.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var internalErr *errs.Error
			teamMemberGroup, internalErr = t.teamMemberGroupRepo.FindMemberGroupByID(ct, tx, id)
			if internalErr != nil {
				return internalErr
			}

			return t.teamMemberGroupDao.DeleteMemberGroup(ct, tx, id)
		})
	return teamMemberGroup, internalErr
}

func (t Team) AddUserToTeamMemberGroup(ct context.Context, memberGroupID uint64, userID uint64) (entity.User, *errs.Error) {
	var user entity.User
	err := t.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var internalErr *errs.Error
			user, internalErr = t.userDao.FindUserByIDWithTx(ct, tx, userID)
			if internalErr != nil {
				return internalErr
			}

			return t.teamMemberGroupUserRelationDao.CreateMemberGroupUserRelation(ct, tx, daoEntity.TeamMemberGroupUserRelation{
				GroupID: memberGroupID,
				UserID:  userID,
			})
		})

	return user, err
}

func (t Team) RemoveUserFromTeamMemberGroup(ct context.Context, memberGroupID uint64, userID uint64) (entity.User, *errs.Error) {
	var user entity.User
	err := t.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var internalErr *errs.Error
			user, internalErr = t.userDao.FindUserByIDWithTx(ct, tx, userID)
			if internalErr != nil {
				return internalErr
			}

			return t.teamMemberGroupUserRelationDao.DeleteMemberGroupUserRelation(ct, tx, daoEntity.TeamMemberGroupUserRelation{
				GroupID: memberGroupID,
				UserID:  userID,
			})
		})

	return user, err
}

func (t Team) MoveUpTeamMemberGroup(
	ct context.Context,
	teamMemberGroupID uint64,
) (entity.TeamMemberGroup, *errs.Error) {
	var teamMemberGroup entity.TeamMemberGroup
	now := time.Now().UTC()
	err := t.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var internalErr *errs.Error
			rawTeamMemberGroup, internalErr := t.teamMemberGroupDao.FindMemberGroupByID(ct, tx, teamMemberGroupID)
			if internalErr != nil {
				return internalErr
			}

			teamMemberGroups, internalErr := t.teamMemberGroupRepo.FindMemberGroupsByTeamID(ct, tx, rawTeamMemberGroup.TeamID)
			if internalErr != nil {
				return internalErr
			}

			currentOrderIndex := rawTeamMemberGroup.Order
			if currentOrderIndex == len(teamMemberGroups)-1 {
				t.logger.WarningWithContext(ct, fmt.Sprintf("team member group is already at the top: teamMemberGroupID=%v", teamMemberGroupID))
				return nil
			}

			nextOrderIndex := currentOrderIndex + 1
			rawTeamMemberGroup.Order = nextOrderIndex
			rawTeamMemberGroup.UpdatedAt = &now
			updateTeamMemberGroupMutation := mutation.NewUpdateTeamMemberGroup(
				t.logger,
				t.stateSyncer,
				t.teamMemberGroupDao,
				rawTeamMemberGroup,
			)

			internalErr = updateTeamMemberGroupMutation.Execute(ct, tx)
			if internalErr != nil {
				return internalErr
			}

			rtTx.AppendMutation(updateTeamMemberGroupMutation)
			nextTeamMemberGroups := collect.Filter(teamMemberGroups, func(group entity.TeamMemberGroup) bool {
				return group.Order == nextOrderIndex
			})
			if len(nextTeamMemberGroups) != 1 {
				return errs.NewError(errs.Unknown, fmt.Sprintf("team member group not found: teamMemberGroupID=%v", teamMemberGroupID))
			}

			nextTeamMemberGroupId := nextTeamMemberGroups[0].ID
			rawNextTeamMemberGroup, internalErr := t.teamMemberGroupDao.FindMemberGroupByID(ct, tx, nextTeamMemberGroupId)
			if internalErr != nil {
				return internalErr
			}

			rawNextTeamMemberGroup.Order = currentOrderIndex
			rawNextTeamMemberGroup.UpdatedAt = &now
			updateNextTeamMemberGroupMutation := mutation.NewUpdateTeamMemberGroup(
				t.logger,
				t.stateSyncer,
				t.teamMemberGroupDao,
				rawNextTeamMemberGroup,
			)

			internalErr = updateNextTeamMemberGroupMutation.Execute(ct, tx)
			if internalErr != nil {
				return internalErr
			}

			rtTx.AppendMutation(updateNextTeamMemberGroupMutation)
			teamMemberGroup = repository.GetTeamMemberGroupFromRawTeamMemberGroup(rawNextTeamMemberGroup)
			return nil
		})

	return teamMemberGroup, err
}

func (t Team) MoveDownTeamMemberGroup(
	ct context.Context,
	teamMemberGroupID uint64,
) (entity.TeamMemberGroup, *errs.Error) {
	var teamMemberGroup entity.TeamMemberGroup
	now := time.Now().UTC()
	err := t.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var internalErr *errs.Error
			rawTeamMemberGroup, internalErr := t.teamMemberGroupDao.FindMemberGroupByID(ct, tx, teamMemberGroupID)
			if internalErr != nil {
				return internalErr
			}

			teamMemberGroups, internalErr := t.teamMemberGroupRepo.FindMemberGroupsByTeamID(ct, tx, rawTeamMemberGroup.TeamID)
			if internalErr != nil {
				return internalErr
			}

			currentOrderIndex := rawTeamMemberGroup.Order
			if currentOrderIndex == 0 {
				t.logger.WarningWithContext(ct, fmt.Sprintf("team member group is already at the bottom: teamMemberGroupID=%v", teamMemberGroupID))
				return nil
			}

			prevOrderIndex := currentOrderIndex - 1
			rawTeamMemberGroup.Order = prevOrderIndex
			rawTeamMemberGroup.UpdatedAt = &now
			updateTeamMemberGroupMutation := mutation.NewUpdateTeamMemberGroup(
				t.logger,
				t.stateSyncer,
				t.teamMemberGroupDao,
				rawTeamMemberGroup,
			)

			internalErr = updateTeamMemberGroupMutation.Execute(ct, tx)
			if internalErr != nil {
				return internalErr
			}

			rtTx.AppendMutation(updateTeamMemberGroupMutation)
			prevTeamMemberGroups := collect.Filter(teamMemberGroups, func(group entity.TeamMemberGroup) bool {
				return group.Order == prevOrderIndex
			})
			if len(prevTeamMemberGroups) != 1 {
				return errs.NewError(errs.Unknown, fmt.Sprintf("team member group not found: teamMemberGroupID=%v", teamMemberGroupID))
			}

			prevTeamMemberGroupId := prevTeamMemberGroups[0].ID
			rawPrevTeamMemberGroup, internalErr := t.teamMemberGroupDao.FindMemberGroupByID(ct, tx, prevTeamMemberGroupId)
			if internalErr != nil {
				return internalErr
			}

			rawPrevTeamMemberGroup.Order = currentOrderIndex
			rawPrevTeamMemberGroup.UpdatedAt = &now
			updatePrevTeamMemberGroupMutation := mutation.NewUpdateTeamMemberGroup(
				t.logger,
				t.stateSyncer,
				t.teamMemberGroupDao,
				rawPrevTeamMemberGroup,
			)

			internalErr = updatePrevTeamMemberGroupMutation.Execute(ct, tx)
			if internalErr != nil {
				return internalErr
			}

			rtTx.AppendMutation(updatePrevTeamMemberGroupMutation)
			teamMemberGroup = repository.GetTeamMemberGroupFromRawTeamMemberGroup(rawPrevTeamMemberGroup)
			return nil
		})

	return teamMemberGroup, err
}

func NewTeam(
	logger telemetry.Logger,
	transactionGroupFactory transaction.GroupFactory,
	cloudWebAPIExternalBaseURL string,
	cloudClientRegistry *client.Registry,
	authorizer client.Authorizer,
	featureToggles feature.Toggles,
	stateSyncer *realtime.StateSyncer,
	transactionFactory cloudTransaction.Factory,
	cache *cache.TimeBasedCache[string, any],
	taskDao dao.Task,
	sprintDao dao.Sprint,
	sprintParticipantDao dao.SprintParticipant,
	teamDao dao.Team,
	userDao dao.User,
	teamMemberDao dao.TeamMember,
	teamFileUploadSessionDao dao.TeamFileUploadSession,
	teamMemberGroupDao dao.TeamMemberGroup,
	teamMemberGroupUserRelationDao dao.TeamMemberGroupUserRelation,
	teamMemberGroupRepo repository.TeamMemberGroup,
) Team {
	return Team{
		logger:                         logger,
		transactionGroupFactory:        transactionGroupFactory,
		cloudWebAPIExternalBaseURL:     cloudWebAPIExternalBaseURL,
		cloudClientRegistry:            cloudClientRegistry,
		authorizer:                     authorizer,
		featureToggles:                 featureToggles,
		stateSyncer:                    stateSyncer,
		transactionFactory:             transactionFactory,
		cache:                          cache,
		taskDao:                        taskDao,
		sprintDao:                      sprintDao,
		sprintParticipantDao:           sprintParticipantDao,
		teamMemberDao:                  teamMemberDao,
		teamDao:                        teamDao,
		userDao:                        userDao,
		teamFileUploadSessionDao:       teamFileUploadSessionDao,
		teamMemberGroupDao:             teamMemberGroupDao,
		teamMemberGroupUserRelationDao: teamMemberGroupUserRelationDao,
		teamMemberGroupRepo:            teamMemberGroupRepo,
	}
}

func findTeamCacheKey(teamID uint64) string {
	return fmt.Sprintf("FindTeamByID(%d)", teamID)
}

func findTeamsCacheKey(filter *TeamFilter) string {
	return fmt.Sprintf("FindTeams(%v)", filter)
}
