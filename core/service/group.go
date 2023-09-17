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
	Name   string
	Filter string
}
type UpdateFilterGroupInput struct {
	Name   string
	Filter string
}

type CreateStaticTeamGroupInput struct {
	Name    string
	TeamIDs []uint64
}

type UpdateStaticTeamGroupInput struct {
	Name    string
	TeamIDs []uint64
}

type CreateStaticUserGroupInput struct {
	Name    string
	UserIDs []uint64
}

type UpdateStaticUserGroupInput struct {
	Name    string
	UserIDs []uint64
}

type Group struct {
	logger               telemetry.Logger
	cloudClientRegistry  *client.Registry
	transactionFactory   cloudTransaction.Factory
	stateSyncer          *realtime.StateSyncer
	groupRepository      repository.Group
	userGroupRelationDao dao.UserGroupRelation
	appGroupRelationDao  dao.AppGroupRelation
	teamGroupRelationDao dao.TeamGroupRelation
	userDao              dao.User
	teamDao              dao.Team
	appDao               dao.App
}

func (g *Group) CreateStaticUserGroup(
	ct context.Context,
	appID uint64,
	input CreateStaticUserGroupInput,
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
			ID:        genGroupIDRes.UniqueNumber,
			Name:      input.Name,
			Type:      entity.GroupTypeStatic,
			CreatedAt: now,
		},
	}
	txCtx := transaction.NewTransactionsContext(
		g.logger,
		g.transactionFactory,
		g.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		err := g.groupRepository.CreateStaticGroup(ct, tx, group)
		if err != nil {
			return err
		}

		appGroupRelation := entity.AppGroupRelation{
			AppID:   appID,
			GroupID: group.ID,
			Type:    entity.AppGroupRelationTypeUser,
		}
		err = g.appGroupRelationDao.CreateAppGroupRelation(ct, tx, appGroupRelation)
		if err != nil {
			return err
		}

		for _, userID := range input.UserIDs {
			err := g.userGroupRelationDao.CreateUserGroupRelation(ct, tx, entity.UserGroupRelation{
				UserID:  userID,
				GroupID: group.ID,
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
	return group, err
}

func (g *Group) UpdateStaticUserGroup(
	ct context.Context,
	groupID uint64,
	input UpdateStaticUserGroupInput,
) (entity.StaticGroup, *errs.Error) {
	now := time.Now().UTC()
	group := entity.StaticGroup{
		Group: entity.Group{
			ID:        groupID,
			Name:      input.Name,
			Type:      entity.GroupTypeStatic,
			CreatedAt: now,
		},
	}
	txCtx := transaction.NewTransactionsContext(
		g.logger,
		g.transactionFactory,
		g.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		err := g.groupRepository.UpdateStaticGroup(ct, tx, group)
		if err != nil {
			return err
		}

		currentUserIDs, err := g.userGroupRelationDao.FindUserIDsByGroupID(ct, group.ID)
		if err != nil {
			return err
		}

		currentUserIDsSet := map[uint64]bool{}
		for _, userID := range currentUserIDs {
			currentUserIDsSet[userID] = true
		}

		userIDsSet := map[uint64]bool{}
		for _, userID := range input.UserIDs {
			userIDsSet[userID] = true
		}

		detected := delta.DetectMapDelta(
			currentUserIDsSet,
			userIDsSet,
			delta.DetectValueDelta[bool],
			delta.ToValueDelta[bool],
		)
		for userID, detectedValue := range detected.Value {
			switch detectedValue.KeyStatus {
			case delta.AddedStatus:
				err := g.userGroupRelationDao.CreateUserGroupRelation(ct, tx, entity.UserGroupRelation{
					UserID:  userID,
					GroupID: group.ID,
				})
				if err != nil {
					return err
				}
			case delta.RemovedStatus:
				err := g.userGroupRelationDao.DeleteUserGroupRelation(ct, tx, group.ID, userID)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
	return group, err
}

func (g *Group) CreateFilterUserGroup(
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
			ID:        genGroupIDRes.UniqueNumber,
			Name:      input.Name,
			Type:      entity.GroupTypeFilter,
			CreatedAt: now,
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
			Type:    entity.AppGroupRelationTypeUser,
		}
		return g.appGroupRelationDao.CreateAppGroupRelation(ct, tx, appGroupRelation)
	})
	return filterGroup, err
}

func (g *Group) UpdateFilterGroup(ct context.Context, groupID uint64, input UpdateFilterGroupInput) (entity.FilterGroup, *errs.Error) {
	now := time.Now().UTC()
	filterGroup := entity.FilterGroup{
		Group: entity.Group{
			ID:        groupID,
			Name:      input.Name,
			Type:      entity.GroupTypeFilter,
			UpdatedAt: &now,
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
		return g.groupRepository.UpdateFilterGroup(ct, tx, filterGroup)
	})
	if err != nil {
		return entity.FilterGroup{}, err
	}

	return filterGroup, err
}

func (g *Group) CreateStaticTeamGroup(
	ct context.Context,
	appID uint64,
	input CreateStaticTeamGroupInput,
) (entity.StaticGroup, *errs.Error) {
	genGroupIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "groupID"}
	genGroupIDRes, rpcErr := g.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genGroupIDReq)
	if rpcErr != nil {
		return entity.StaticGroup{}, errs.FromGRPCErr(rpcErr)
	}

	now := time.Now().UTC()
	group := entity.StaticGroup{
		Group: entity.Group{
			ID:        genGroupIDRes.UniqueNumber,
			Name:      input.Name,
			Type:      entity.GroupTypeStatic,
			CreatedAt: now,
		},
	}
	txCtx := transaction.NewTransactionsContext(
		g.logger,
		g.transactionFactory,
		g.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		err := g.groupRepository.CreateStaticGroup(ct, tx, group)
		if err != nil {
			return err
		}

		appGroupRelation := entity.AppGroupRelation{
			AppID:   appID,
			GroupID: group.ID,
			Type:    entity.AppGroupRelationTypeTeam,
		}
		err = g.appGroupRelationDao.CreateAppGroupRelation(ct, tx, appGroupRelation)
		if err != nil {
			return err
		}

		for _, teamID := range input.TeamIDs {
			err := g.teamGroupRelationDao.CreateTeamGroupRelation(ct, tx, entity.TeamGroupRelation{
				TeamID:  teamID,
				GroupID: group.ID,
			})
			if err != nil {
				return err
			}
		}
		
		return nil
	})
	return group, err
}

func (g *Group) UpdateStaticTeamGroup(
	ct context.Context,
	groupID uint64,
	input UpdateStaticTeamGroupInput,
) (entity.StaticGroup, *errs.Error) {
	now := time.Now().UTC()
	group := entity.StaticGroup{
		Group: entity.Group{
			ID:        groupID,
			Name:      input.Name,
			Type:      entity.GroupTypeStatic,
			CreatedAt: now,
		},
	}
	txCtx := transaction.NewTransactionsContext(
		g.logger,
		g.transactionFactory,
		g.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		err := g.groupRepository.UpdateStaticGroup(ct, tx, group)
		if err != nil {
			return err
		}

		currentTeamIDs, err := g.teamGroupRelationDao.FindTeamIDsByGroupID(ct, group.ID)
		if err != nil {
			return err
		}

		currentTeamIDsSet := map[uint64]bool{}
		for _, teamID := range currentTeamIDs {
			currentTeamIDsSet[teamID] = true
		}

		teamIDsSet := map[uint64]bool{}
		for _, teamID := range input.TeamIDs {
			teamIDsSet[teamID] = true
		}

		detected := delta.DetectMapDelta(
			currentTeamIDsSet,
			teamIDsSet,
			delta.DetectValueDelta[bool],
			delta.ToValueDelta[bool],
		)
		for teamID, detectedValue := range detected.Value {
			switch detectedValue.KeyStatus {
			case delta.AddedStatus:
				err := g.teamGroupRelationDao.CreateTeamGroupRelation(ct, tx, entity.TeamGroupRelation{
					TeamID:  teamID,
					GroupID: group.ID,
				})
				if err != nil {
					return err
				}
			case delta.RemovedStatus:
				err := g.teamGroupRelationDao.DeleteTeamGroupRelation(ct, tx, group.ID, teamID)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
	return group, err
}

func (g *Group) CreateFilterTeamGroup(
	ct context.Context,
	appID uint64,
	input CreateFilterGroupInput,
) (entity.FilterGroup, *errs.Error) {
	genGroupIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "groupID"}
	genGroupIDRes, rpcErr := g.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genGroupIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.FilterGroup{}, internalErr
	}

	now := time.Now().UTC()
	filterGroup := entity.FilterGroup{
		Group: entity.Group{
			ID:        genGroupIDRes.UniqueNumber,
			Name:      input.Name,
			Type:      entity.GroupTypeFilter,
			CreatedAt: now,
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
			Type:    entity.AppGroupRelationTypeTeam,
		}
		return g.appGroupRelationDao.CreateAppGroupRelation(ct, tx, appGroupRelation)
	})
	return filterGroup, err
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
		userIDs, err := g.userGroupRelationDao.FindUserIDsByGroupID(ct, groupID)
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
		teamIDs, err := g.teamGroupRelationDao.FindTeamIDsByGroupID(ct, groupID)
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

func (g *Group) FindAppByGroupID(ct context.Context, groupID uint64) (entity.App, *errs.Error) {
	appID, err := g.appGroupRelationDao.FindAppIDByGroupID(ct, groupID)
	if err != nil {
		return entity.App{}, err
	}

	return g.appDao.FindAppByID(ct, appID)
}

func (g *Group) FindUserGroupsByAppID(ct context.Context, appID uint64) ([]entity.GroupUnion, *errs.Error) {
	return g.findGroupsByAppID(ct, appID, entity.AppGroupRelationTypeUser)
}

func (g *Group) FindTeamGroupsByAppID(ct context.Context, appID uint64) ([]entity.GroupUnion, *errs.Error) {
	return g.findGroupsByAppID(ct, appID, entity.AppGroupRelationTypeTeam)
}

func (g *Group) findGroupsByAppID(ct context.Context, appID uint64, appGroupRelationType entity.AppGroupRelationType) ([]entity.GroupUnion, *errs.Error) {
	groupIDs, err := g.appGroupRelationDao.FindGroupIDsByAppIDAndRelationType(ct, appID, appGroupRelationType)
	if err != nil {
		return nil, err
	}

	return g.groupRepository.FindGroupsByIDs(ct, groupIDs)
}

func NewGroup(
	logger telemetry.Logger,
	cloudClientRegistry *client.Registry,
	transactionFactory cloudTransaction.Factory,
	stateSyncer *realtime.StateSyncer,
	userGroupRelationDao dao.UserGroupRelation,
	appGroupRelationDao dao.AppGroupRelation,
	teamGroupRelationDao dao.TeamGroupRelation,
	userDao dao.User,
	teamDao dao.Team,
	appDao dao.App,
) *Group {
	return &Group{
		logger:               logger,
		cloudClientRegistry:  cloudClientRegistry,
		transactionFactory:   transactionFactory,
		stateSyncer:          stateSyncer,
		userGroupRelationDao: userGroupRelationDao,
		appGroupRelationDao:  appGroupRelationDao,
		teamGroupRelationDao: teamGroupRelationDao,
		userDao:              userDao,
		teamDao:              teamDao,
		appDao:               appDao,
	}
}
