package service

import (
	"context"
	"time"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/delta"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudTransaction "github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/repository"
	"github.com/teamyapp/teamy-backend/core/transaction"
)

type CreateFilterGroupInput struct {
	Name            string
	Filter          string
	GroupMemberType entity.GroupMemberType
	RolloutIDs      []uint64
}

type CreateStaticGroupInput struct {
	Name            string
	MemberIDs       []uint64
	GroupMemberType entity.GroupMemberType
	RolloutIDs      []uint64
}

type UpdateGroupInput struct {
	Name            string
	Filter          string
	GroupMemberType entity.GroupMemberType
	RolloutIDs      []uint64
	MemberIDs       []uint64
	Type            entity.GroupType
}

type Group struct {
	logger                  telemetry.Logger
	cloudClientRegistry     *client.Registry
	transactionFactory      cloudTransaction.Factory
	stateSyncer             *realtime.StateSyncer
	groupRepository         *repository.Group
	appGroupRelationDao     dao.AppGroupRelation
	groupMemberRelation     dao.GroupMemberRelation
	groupRolloutRelationDao dao.GroupRolloutRelation
	userDao                 dao.User
	teamDao                 dao.Team
	appDao                  dao.App
	groupDao                dao.Group
	appRolloutRelation      dao.AppRolloutRelation
}

func (g *Group) CreateAppStaticGroup(
	ct context.Context,
	appID uint64,
	input CreateStaticGroupInput,
) (entity.StaticGroup, *errs.Error) {
	genGroupIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "groupID"}
	genGroupIDRes, rpcErr := g.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genGroupIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.StaticGroup{}, internalErr
	}

	now := time.Now().UTC()
	group := entity.StaticGroup{
		Group: entity.Group{
			ID:              genGroupIDRes.UniqueNumber,
			Name:            input.Name,
			Type:            entity.GroupTypeStatic,
			MemberType:      input.GroupMemberType,
			MaxRolloutIndex: len(input.RolloutIDs) - 1,
			CreatedAt:       now,
		},
	}
	txCtx := transaction.NewTransactionsContext(
		g.logger,
		g.transactionFactory,
		g.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		err := g.groupRepository.CreateStaticGroup(ct, tx, group)
		if err != nil {
			return err
		}

		appGroupRelation := entity.AppGroupRelation{
			AppID:   appID,
			GroupID: group.ID,
		}
		err = g.appGroupRelationDao.CreateAppGroupRelation(ct, tx, appGroupRelation)
		if err != nil {
			return err
		}

		for _, memberID := range input.MemberIDs {
			switch input.GroupMemberType {
			case entity.GroupMemberTypeUser:
				_, err := g.userDao.FindUserByIDWithTx(ct, tx, memberID)
				if err != nil {
					return err
				}

			case entity.GroupMemberTypeTeam:
				_, err := g.teamDao.FindTeamByIDWithTx(ct, tx, memberID)
				if err != nil {
					return err
				}

			default:
				return errs.NewError(errs.InvalidArgument, "invalid group member type")
			}

			err = g.groupMemberRelation.CreateGroupMemberRelation(ct, tx, entity.GroupMemberRelation{
				MemberID: memberID,
				GroupID:  group.ID,
			})
			if err != nil {
				return err
			}
		}

		for index, rolloutID := range input.RolloutIDs {
			err := g.groupRolloutRelationDao.CreateGroupRolloutRelation(ct, tx, entity.GroupRolloutRelation{
				GroupID:    group.ID,
				RolloutID:  rolloutID,
				OrderIndex: index,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return group, err
}

func (g *Group) UpdateGroup(ct context.Context, appID uint64, groupID uint64, input UpdateGroupInput) (entity.GroupUnion, *errs.Error) {
	var groupUnion entity.GroupUnion
	txCtx := transaction.NewTransactionsContext(
		g.logger,
		g.transactionFactory,
		g.stateSyncer,
		ct,
	)

	now := time.Now().UTC()
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		group, err := g.groupDao.FindGroupByIDWithTx(ct, tx, groupID)
		if err != nil {
			return err
		}

		// Currently, we do not support update group member type
		switch group.MemberType {
		case entity.GroupMemberTypeUser:
			if input.GroupMemberType != entity.GroupMemberTypeUser {
				return errs.NewError(errs.InvalidArgument, "invalid group member type")
			}
		case entity.GroupMemberTypeTeam:
			if input.GroupMemberType != entity.GroupMemberTypeTeam {
				return errs.NewError(errs.InvalidArgument, "invalid group member type")
			}
		default:
			return errs.NewError(errs.InvalidArgument, "invalid group member type")
		}

		switch input.Type {
		case entity.GroupTypeStatic:
			err = g.updateGroupMemberRelations(ct, tx, groupID, input.MemberIDs, input.GroupMemberType)
		case entity.GroupTypeFilter:
			break
		default:
			err = errs.NewError(errs.InvalidArgument, "invalid group type")
		}

		if err != nil {
			return err
		}

		currentMaxRolloutIndex, err := g.updateGroupRolloutRelations(
			ct,
			tx,
			groupID,
			appID,
			input.GroupMemberType,
			input.RolloutIDs,
			group.MaxRolloutIndex,
		)
		if err != nil {
			return err
		}

		updatedGroup := entity.Group{
			ID:              groupID,
			Type:            input.Type,
			MemberType:      input.GroupMemberType,
			UpdatedAt:       &now,
			CreatedAt:       group.CreatedAt,
			Name:            input.Name,
			MaxRolloutIndex: currentMaxRolloutIndex,
		}

		if group.Type != input.Type {
			err = g.groupRepository.DeletePartialGroup(ct, tx, groupID)
			if err != nil {
				return err
			}

			partialGroupInput := repository.CreatePartialGroupInput{
				ID:     groupID,
				Filter: input.Filter,
				Type:   input.Type,
			}
			err = g.groupRepository.CreatePartialGroup(ct, tx, partialGroupInput)
			if err != nil {
				return err
			}

			err = g.groupDao.UpdateGroup(ct, tx, updatedGroup)
			if err != nil {
				return err
			}

			groupUnion, err = g.groupRepository.GetGroupUnionFromBaseGroup(ct, tx, updatedGroup)
			return err
		}

		groupUnion.Type = input.Type
		groupUnion.MemberType = input.GroupMemberType
		switch input.Type {
		case entity.GroupTypeStatic:
			groupUnion.StaticGroup = entity.StaticGroup{
				Group: updatedGroup,
			}

			err = g.groupRepository.UpdateStaticGroup(ct, tx, groupUnion.StaticGroup)
		case entity.GroupTypeFilter:
			groupUnion.FilterGroup = entity.FilterGroup{
				Group:  updatedGroup,
				Filter: input.Filter,
				Count:  0,
			}

			err = g.groupRepository.UpdateFilterGroup(ct, tx, groupUnion.FilterGroup)
		default:
			err = errs.NewError(errs.InvalidArgument, "invalid group type")
		}

		return err
	})

	return groupUnion, err
}

func (g *Group) FindUsersByGroupID(ct context.Context, groupID uint64) ([]entity.User, *errs.Error) {
	var users []entity.User
	txCtx := transaction.NewTransactionsContext(
		g.logger,
		g.transactionFactory,
		g.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		userIDs, err := g.groupMemberRelation.FindMemberIDsByGroupIDWithTx(ct, tx, groupID)
		if err != nil {
			return err
		}

		users, err = g.userDao.FindUsersByIDsWithTx(ct, tx, userIDs)
		if err != nil {
			return err
		}

		return nil
	})
	return users, err
}

func (g *Group) FindTeamsByGroupID(ct context.Context, groupID uint64) ([]entity.Team, *errs.Error) {
	var teams []entity.Team
	txCtx := transaction.NewTransactionsContext(
		g.logger,
		g.transactionFactory,
		g.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		teamIDs, err := g.groupMemberRelation.FindMemberIDsByGroupIDWithTx(ct, tx, groupID)
		if err != nil {
			return err
		}

		teams, err = g.teamDao.FindTeamsByIDsWithTx(ct, tx, teamIDs)
		if err != nil {
			return err
		}

		return nil
	})
	return teams, err
}

func (g *Group) FindGroupsByAppID(ct context.Context, appID uint64) ([]entity.GroupUnion, *errs.Error) {
	var groups []entity.GroupUnion
	txCtx := transaction.NewTransactionsContext(
		g.logger,
		g.transactionFactory,
		g.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		groupIDs, err := g.appGroupRelationDao.FindGroupIDsByAppIDWithTx(ct, tx, appID)
		if err != nil {
			return err
		}

		groups, err = g.groupRepository.FindGroupsByIDsWithTx(ct, tx, groupIDs)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (g *Group) FindGroupRolloutRelationsByGroupID(ct context.Context, groupID uint64) ([]entity.GroupRolloutRelation, *errs.Error) {
	return g.groupRolloutRelationDao.FindGroupRolloutRelationsByGroupID(ct, groupID)
}

func (g *Group) FindGroupByID(ct context.Context, groupID uint64) (entity.GroupUnion, *errs.Error) {
	var group entity.GroupUnion
	txCtx := transaction.NewTransactionsContext(
		g.logger,
		g.transactionFactory,
		g.stateSyncer,
		ct,
	)

	err := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		group, err = g.groupRepository.FindGroupByIDWithTx(ct, tx, groupID)
		return err
	})

	return group, err
}

func (g *Group) DeleteAppGroup(ct context.Context, appID uint64, groupID uint64) (entity.GroupUnion, *errs.Error) {
	var group entity.GroupUnion
	txCtx := transaction.NewTransactionsContext(
		g.logger,
		g.transactionFactory,
		g.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		err = g.appGroupRelationDao.DeleteAppGroupRelation(ct, tx, appID, groupID)
		if err != nil {
			return err
		}

		err = g.groupMemberRelation.DeleteGroupMemberRelationsByGroupID(ct, tx, groupID)
		if err != nil {
			return err
		}

		err = g.groupRolloutRelationDao.DeleteGroupRolloutRelationsByGroupID(ct, tx, groupID)
		if err != nil {
			return err
		}

		group, err = g.groupRepository.DeleteGroup(ct, tx, groupID)
		if err != nil {
			return err
		}

		return nil
	})

	return group, err
}

func (g *Group) CreateFilterGroup(
	ct context.Context,
	appID uint64,
	input CreateFilterGroupInput,
) (entity.FilterGroup, *errs.Error) {
	genGroupIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "groupID"}
	genGroupIDRes, rpcErr := g.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genGroupIDReq)
	if rpcErr != nil {
		return entity.FilterGroup{}, errs.FromGRPCErr(rpcErr)
	}

	now := time.Now().UTC()
	filterGroup := entity.FilterGroup{
		Group: entity.Group{
			ID:              genGroupIDRes.UniqueNumber,
			Name:            input.Name,
			Type:            entity.GroupTypeFilter,
			MemberType:      input.GroupMemberType,
			MaxRolloutIndex: len(input.RolloutIDs) - 1,
			CreatedAt:       now,
		},
		Filter: input.Filter,
		Count:  0,
	}
	txCtx := transaction.NewTransactionsContext(
		g.logger,
		g.transactionFactory,
		g.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		err := g.groupRepository.CreateFilterGroup(ct, tx, filterGroup)
		if err != nil {
			return err
		}

		for index, rolloutID := range input.RolloutIDs {
			err := g.groupRolloutRelationDao.CreateGroupRolloutRelation(ct, tx, entity.GroupRolloutRelation{
				GroupID:    filterGroup.ID,
				RolloutID:  rolloutID,
				OrderIndex: index,
			})
			if err != nil {
				return err
			}
		}

		appGroupRelation := entity.AppGroupRelation{
			AppID:   appID,
			GroupID: filterGroup.ID,
		}
		err = g.appGroupRelationDao.CreateAppGroupRelation(ct, tx, appGroupRelation)
		return err
	})

	return filterGroup, err
}

func (g *Group) updateGroupMemberRelations(
	ct context.Context,
	tx *cloudTransaction.Transaction,
	groupID uint64,
	memberIDs []uint64,
	groupMemberType entity.GroupMemberType,
) *errs.Error {
	memberIDsSet := map[uint64]bool{}
	for _, memberID := range memberIDs {
		switch groupMemberType {
		case entity.GroupMemberTypeUser:
			_, err := g.userDao.FindUserByIDWithTx(ct, tx, memberID)
			if err != nil {
				return err
			}
		case entity.GroupMemberTypeTeam:
			_, err := g.teamDao.FindTeamByIDWithTx(ct, tx, memberID)
			if err != nil {
				return err
			}
		default:
			return errs.NewError(errs.InvalidArgument, "invalid group member type")
		}

		memberIDsSet[memberID] = true
	}

	currentMemberIDs, err := g.groupMemberRelation.FindMemberIDsByGroupIDWithTx(ct, tx, groupID)
	if err != nil {
		return err
	}

	currentMemberIDsSet := map[uint64]bool{}
	for _, userID := range currentMemberIDs {
		currentMemberIDsSet[userID] = true
	}

	detected := delta.DetectMapDelta(
		currentMemberIDsSet,
		memberIDsSet,
		delta.DetectValueDelta[bool],
		delta.ToValueDelta[bool],
	)
	for memberID, detectedValue := range detected.Value {
		switch detectedValue.KeyStatus {
		case delta.AddedStatus:
			err := g.groupMemberRelation.CreateGroupMemberRelation(ct, tx, entity.GroupMemberRelation{
				MemberID: memberID,
				GroupID:  groupID,
			})
			if err != nil {
				return err
			}
		case delta.RemovedStatus:
			err := g.groupMemberRelation.DeleteGroupMemberRelation(ct, tx, memberID, groupID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Group) updateGroupRolloutRelations(
	ct context.Context,
	tx *cloudTransaction.Transaction,
	groupID uint64,
	appID uint64,
	groupMemberType entity.GroupMemberType,
	rolloutIDs []uint64,
	currentMaxRolloutIndex int,
) (int, *errs.Error) {
	currentRolloutIDs, err := g.groupRolloutRelationDao.FindGroupRolloutRelationsByGroupIDWithTx(ct, tx, groupID)
	if err != nil {
		return 0, err
	}

	currentRolloutIDsSet := map[uint64]bool{}
	for _, groupRolloutRelation := range currentRolloutIDs {
		currentRolloutIDsSet[groupRolloutRelation.RolloutID] = true
	}

	rolloutIDsSet := map[uint64]bool{}
	for _, rolloutID := range rolloutIDs {
		appRolloutRelation, err := g.appRolloutRelation.FindAppRolloutByAppIDAndRolloutIDWithTx(ct, tx, appID, rolloutID)
		if err != nil {
			return 0, err
		}

		if appRolloutRelation.Type != entity.AppRolloutRelationType(groupMemberType) {
			return 0, errs.NewError(errs.InvalidArgument, "invalid rollout type")
		}

		rolloutIDsSet[rolloutID] = true
	}

	detected := delta.DetectMapDelta(
		currentRolloutIDsSet,
		rolloutIDsSet,
		delta.DetectValueDelta[bool],
		delta.ToValueDelta[bool],
	)

	for rolloutID, detectedValue := range detected.Value {
		switch detectedValue.KeyStatus {
		case delta.AddedStatus:
			err := g.groupRolloutRelationDao.CreateGroupRolloutRelation(ct, tx, entity.GroupRolloutRelation{
				GroupID:    groupID,
				RolloutID:  rolloutID,
				OrderIndex: currentMaxRolloutIndex + 1,
			})
			if err != nil {
				return 0, err
			}

			currentMaxRolloutIndex++
		case delta.RemovedStatus:
			relation, err := g.groupRolloutRelationDao.FindGroupRolloutByGroupIDAndRolloutIDWithTx(ct, tx, groupID, rolloutID)
			if err != nil {
				return 0, err
			}

			err = g.groupRolloutRelationDao.DeleteGroupRolloutRelationsByGroupIDAndRolloutID(ct, tx, groupID, rolloutID)
			if err != nil {
				return 0, err
			}

			err = g.groupRolloutRelationDao.UpdateFromOrderIndexByGroupID(ct, tx, -1, relation.OrderIndex+1, groupID)
			if err != nil {
				return 0, err
			}

			currentMaxRolloutIndex--
		}
	}

	return currentMaxRolloutIndex, nil
}

func NewGroup(
	logger telemetry.Logger,
	cloudClientRegistry *client.Registry,
	transactionFactory cloudTransaction.Factory,
	stateSyncer *realtime.StateSyncer,
	groupRepository *repository.Group,
	groupMemberRelation dao.GroupMemberRelation,
	appGroupRelationDao dao.AppGroupRelation,
	groupRolloutRelationDao dao.GroupRolloutRelation,
	userDao dao.User,
	teamDao dao.Team,
	appDao dao.App,
	groupDao dao.Group,
	appRolloutRelation dao.AppRolloutRelation,
) *Group {
	return &Group{
		logger:                  logger,
		cloudClientRegistry:     cloudClientRegistry,
		transactionFactory:      transactionFactory,
		stateSyncer:             stateSyncer,
		groupRepository:         groupRepository,
		groupMemberRelation:     groupMemberRelation,
		appGroupRelationDao:     appGroupRelationDao,
		groupRolloutRelationDao: groupRolloutRelationDao,
		userDao:                 userDao,
		teamDao:                 teamDao,
		appDao:                  appDao,
		groupDao:                groupDao,
		appRolloutRelation:      appRolloutRelation,
	}
}
