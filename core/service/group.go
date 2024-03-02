package service

import (
	"context"
	"fmt"
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
	GroupMemberType entity.GroupMemberType
	Filter          string
	RolloutIDs      []uint64
}

type CreateStaticGroupInput struct {
	Name            string
	GroupMemberType entity.GroupMemberType
	MemberIDs       []uint64
	RolloutIDs      []uint64
}

type UpdateGroupInput struct {
	Name            string
	Type            entity.GroupType
	GroupMemberType entity.GroupMemberType
	Filter          *string
	MemberIDs       []uint64
	RolloutIDs      []uint64
}

type Group struct {
	logger                  telemetry.Logger
	cloudClientRegistry     *client.Registry
	transactionFactory      cloudTransaction.Factory
	stateSyncer             *realtime.StateSyncer
	groupRepository         *repository.Group
	appGroupRelationDao     dao.AppGroupRelation
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
		MemberIDs: input.MemberIDs,
	}
	txCtx := transaction.NewTransactionsContext(
		g.logger,
		g.transactionFactory,
		g.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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
		}

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

		if group.Locked {
			return errs.NewError(errs.InvalidArgument, "group is locked")
		}

		// Currently, we do not support update group member type
		if group.MemberType != input.GroupMemberType {
			return errs.NewError(errs.InvalidArgument, fmt.Sprintf("invalid group member type, current: %s, new: %s", group.MemberType, input.GroupMemberType))
		}

		switch input.Type {
		case entity.GroupTypeStatic:
			err = g.updateStaticGroupMemberRelations(ct, tx, groupID, input.MemberIDs, input.GroupMemberType)
		case entity.GroupTypeFilter:
			break
		default:
			err = errs.NewError(errs.InvalidArgument, fmt.Sprintf("invalid group type %s", input.Type))
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

		var filter string = ""
		if input.Filter != nil {
			filter = *input.Filter
		}

		if group.Type != input.Type {
			err = g.groupRepository.DeletePartialGroup(ct, tx, groupID)
			if err != nil {
				return err
			}

			partialGroupInput := repository.CreatePartialGroupInput{
				ID:     groupID,
				Filter: filter,
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
				Filter: filter,
			}

			err = g.groupRepository.UpdateFilterGroup(ct, tx, groupUnion.FilterGroup)
		default:
			err = errs.NewError(errs.InvalidArgument, "invalid group type")
		}

		return err
	})

	return groupUnion, err
}

func (g *Group) FindUsersByStaticGroupID(ct context.Context, groupID uint64) ([]entity.User, *errs.Error) {
	var users []entity.User
	txCtx := transaction.NewTransactionsContext(
		g.logger,
		g.transactionFactory,
		g.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		groupUnion, err := g.groupRepository.FindGroupByIDWithTx(ct, tx, groupID)
		if err != nil {
			return err
		}

		users, err = g.userDao.FindUsersByIDsWithTx(ct, tx, groupUnion.StaticGroup.MemberIDs)
		if err != nil {
			return err
		}

		return nil
	})
	return users, err
}

func (g *Group) FindTeamsByStaticGroupID(ct context.Context, groupID uint64) ([]entity.Team, *errs.Error) {
	var teams []entity.Team
	txCtx := transaction.NewTransactionsContext(
		g.logger,
		g.transactionFactory,
		g.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		groupUnion, err := g.groupRepository.FindGroupByIDWithTx(ct, tx, groupID)
		if err != nil {
			return err
		}

		teams, err = g.teamDao.FindTeamsByIDsWithTx(ct, tx, groupUnion.StaticGroup.MemberIDs)
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
	return groups, err
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

func (g *Group) DeleteGroup(ct context.Context, groupID uint64) (entity.GroupUnion, *errs.Error) {
	var groupUnion entity.GroupUnion
	txCtx := transaction.NewTransactionsContext(
		g.logger,
		g.transactionFactory,
		g.stateSyncer,
		ct,
	)
	transactionErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		groupUnion, err = g.deleteGroup(ct, tx, groupID, false)
		return err
	})

	return groupUnion, transactionErr
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
		return g.appGroupRelationDao.CreateAppGroupRelation(ct, tx, appGroupRelation)
	})

	return filterGroup, err
}

func (g *Group) updateStaticGroupMemberRelations(
	ct context.Context,
	tx *cloudTransaction.Transaction,
	groupID uint64,
	memberIDs []uint64,
	groupMemberType entity.GroupMemberType,
) *errs.Error {
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
	}

	return g.groupRepository.UpdateStaticGroupMemberRelations(ct, tx, groupID, memberIDs)
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

func (g *Group) deleteGroup(
	ct context.Context,
	tx *cloudTransaction.Transaction,
	groupID uint64,
	deleteLocked bool,
) (entity.GroupUnion, *errs.Error) {
	group, err := g.groupDao.FindGroupByIDWithTx(ct, tx, groupID)
	if err != nil {
		return entity.GroupUnion{}, err
	}

	if group.Locked && !deleteLocked {
		return entity.GroupUnion{}, errs.NewError(errs.InvalidOperation, "group is locked")
	}

	groupUnion, err := g.groupRepository.GetGroupUnionFromBaseGroup(ct, tx, group)
	if err != nil {
		return entity.GroupUnion{}, err
	}

	err = g.appGroupRelationDao.DeleteAppGroupRelationsByGroupID(ct, tx, groupID)
	if err != nil {
		return entity.GroupUnion{}, err
	}

	err = g.groupRolloutRelationDao.DeleteGroupRolloutRelationsByGroupID(ct, tx, groupID)
	if err != nil {
		return entity.GroupUnion{}, err
	}

	err = g.groupRepository.DeleteGroup(ct, tx, group.ID, group.Type)
	return groupUnion, err
}

func NewGroup(
	logger telemetry.Logger,
	cloudClientRegistry *client.Registry,
	transactionFactory cloudTransaction.Factory,
	stateSyncer *realtime.StateSyncer,
	groupRepository *repository.Group,
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
		appGroupRelationDao:     appGroupRelationDao,
		groupRolloutRelationDao: groupRolloutRelationDao,
		userDao:                 userDao,
		teamDao:                 teamDao,
		appDao:                  appDao,
		groupDao:                groupDao,
		appRolloutRelation:      appRolloutRelation,
	}
}
