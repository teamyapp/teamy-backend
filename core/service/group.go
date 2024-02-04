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
}
type UpdateFilterGroupInput struct {
	Name            string
	Filter          string
	GroupMemberType entity.GroupMemberType
}

type CreateStaticGroupInput struct {
	Name            string
	MemberIDs       []uint64
	GroupMemberType entity.GroupMemberType
}

type UpdateStaticGroupInput struct {
	Name            string
	MemberIDs       []uint64
	GroupMemberType entity.GroupMemberType
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
			ID:         genGroupIDRes.UniqueNumber,
			Name:       input.Name,
			Type:       entity.GroupTypeStatic,
			MemberType: entity.GroupMemberTypeUser,
			CreatedAt:  now,
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

		return nil
	})
	return group, err
}

func (g *Group) UpdateStaticGroup(
	ct context.Context,
	groupID uint64,
	input UpdateStaticGroupInput,
) (entity.StaticGroup, *errs.Error) {
	var group entity.GroupUnion
	txCtx := transaction.NewTransactionsContext(
		g.logger,
		g.transactionFactory,
		g.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		group, err = g.groupRepository.FindGroupByIDWithTx(ct, tx, groupID)
		if err != nil {
			return err
		}

		switch input.GroupMemberType {
		case entity.GroupMemberTypeUser:
			if group.MemberType != entity.GroupMemberTypeUser {
				return errs.NewError(errs.InvalidArgument, "group is not user group")
			}
		case entity.GroupMemberTypeTeam:
			if group.MemberType != entity.GroupMemberTypeTeam {
				return errs.NewError(errs.InvalidArgument, "group is not team group")
			}
		default:
			return errs.NewError(errs.InvalidArgument, "invalid group member type")
		}

		if group.Type != entity.GroupTypeStatic {
			return errs.NewError(errs.InvalidArgument, "group is not static group")
		}

		group.StaticGroup.Name = input.Name
		now := time.Now().UTC()
		group.StaticGroup.UpdatedAt = &now
		err = g.groupRepository.UpdateStaticGroup(ct, tx, group.StaticGroup)
		if err != nil {
			return err
		}

		memberIDsSet := map[uint64]bool{}
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

			memberIDsSet[memberID] = true
		}

		currentUserIDs, err := g.groupMemberRelation.FindMemberIDsByGroupIDWithTx(ct, tx, group.StaticGroup.ID)
		if err != nil {
			return err
		}

		currentUserIDsSet := map[uint64]bool{}
		for _, userID := range currentUserIDs {
			currentUserIDsSet[userID] = true
		}

		detected := delta.DetectMapDelta(
			currentUserIDsSet,
			memberIDsSet,
			delta.DetectValueDelta[bool],
			delta.ToValueDelta[bool],
		)
		for memberID, detectedValue := range detected.Value {
			switch detectedValue.KeyStatus {
			case delta.AddedStatus:
				err := g.groupMemberRelation.CreateGroupMemberRelation(ct, tx, entity.GroupMemberRelation{
					MemberID: memberID,
					GroupID:  group.StaticGroup.ID,
				})
				if err != nil {
					return err
				}
			case delta.RemovedStatus:
				err := g.groupMemberRelation.DeleteGroupMemberRelation(ct, tx, group.StaticGroup.ID, memberID)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
	return group.StaticGroup, err
}

func (g *Group) UpdateFilterGroup(ct context.Context, groupID uint64, input UpdateFilterGroupInput) (entity.FilterGroup, *errs.Error) {
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
		if err != nil {
			return err
		}

		if group.Type != entity.GroupTypeFilter {
			return errs.NewError(errs.InvalidArgument, "group is not filter group")
		}

		group.FilterGroup.Name = input.Name
		group.FilterGroup.Filter = input.Filter
		now := time.Now().UTC()
		group.FilterGroup.UpdatedAt = &now

		return g.groupRepository.UpdateFilterGroup(ct, tx, group.FilterGroup)
	})
	if err != nil {
		return entity.FilterGroup{}, err
	}

	return group.FilterGroup, nil
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
	return g.groupRepository.FindGroupByID(ct, groupID)
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
			ID:         genGroupIDRes.UniqueNumber,
			Name:       input.Name,
			Type:       entity.GroupTypeFilter,
			MemberType: input.GroupMemberType,
			CreatedAt:  now,
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
	err := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		err := g.groupRepository.CreateFilterGroup(ct, tx, filterGroup)
		if err != nil {
			return err
		}

		appGroupRelation := entity.AppGroupRelation{
			AppID:   appID,
			GroupID: filterGroup.ID,
		}
		return g.appGroupRelationDao.CreateAppGroupRelation(ct, tx, appGroupRelation)
	})
	return filterGroup, err
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
	}
}
