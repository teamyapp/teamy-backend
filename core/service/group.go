package service

import (
	"context"
	"time"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/delta"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
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
	transactionFactory   transaction.Factory
	stateSyncer          *realtime.StateSyncer
	staticGroupDao       dao.StaticGroup
	filterGroupDao       dao.FilterGroup
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

	//TODO: add transaction
	group, err := g.staticGroupDao.CreateStaticGroup(ct, group)
	if err != nil {
		return entity.StaticGroup{}, err
	}

	appGroupRelation := entity.AppGroupRelation{
		AppID:   appID,
		GroupID: group.ID,
		Type:    entity.AppGroupRelationTypeUser,
	}
	_, err = g.appGroupRelationDao.CreateAppGroupRelation(ct, appGroupRelation)
	if err != nil {
		return entity.StaticGroup{}, err
	}

	for _, userID := range input.UserIDs {
		_, err := g.userGroupRelationDao.CreateUserGroupRelation(ct, entity.UserGroupRelation{
			UserID:  userID,
			GroupID: group.ID,
		})
		if err != nil {
			return entity.StaticGroup{}, err
		}
	}

	return group, nil
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

	//TODO: add transaction
	err := g.staticGroupDao.UpdateStaticGroup(ct, group)
	if err != nil {
		return entity.StaticGroup{}, err
	}

	currentUserIDs, err := g.userGroupRelationDao.FindUserIDsByGroupID(ct, group.ID)
	if err != nil {
		return entity.StaticGroup{}, err
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
			_, err := g.userGroupRelationDao.CreateUserGroupRelation(ct, entity.UserGroupRelation{
				UserID:  userID,
				GroupID: group.ID,
			})
			if err != nil {
				return entity.StaticGroup{}, err
			}
		case delta.RemovedStatus:
			err := g.userGroupRelationDao.DeleteUserGroupRelation(ct, group.ID, userID)
			if err != nil {
				return entity.StaticGroup{}, err
			}
		}
	}

	return group, nil
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

	//TODO: add transaction
	filterGroup, err := g.filterGroupDao.CreateFilterGroup(ct, filterGroup)
	if err != nil {
		return entity.FilterGroup{}, err
	}

	appGroupRelation := entity.AppGroupRelation{
		AppID:   appID,
		GroupID: filterGroup.ID,
		Type:    entity.AppGroupRelationTypeUser,
	}
	_, err = g.appGroupRelationDao.CreateAppGroupRelation(ct, appGroupRelation)
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
	err := g.filterGroupDao.UpdateFilterGroup(ct, filterGroup)
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

	//TODO: add transaction
	group, err := g.staticGroupDao.CreateStaticGroup(ct, group)
	if err != nil {
		return entity.StaticGroup{}, err
	}

	appGroupRelation := entity.AppGroupRelation{
		AppID:   appID,
		GroupID: group.ID,
		Type:    entity.AppGroupRelationTypeTeam,
	}
	_, err = g.appGroupRelationDao.CreateAppGroupRelation(ct, appGroupRelation)
	if err != nil {
		return entity.StaticGroup{}, err
	}

	for _, teamID := range input.TeamIDs {
		_, err := g.teamGroupRelationDao.CreateTeamGroupRelation(ct, entity.TeamGroupRelation{
			TeamID:  teamID,
			GroupID: group.ID,
		})
		if err != nil {
			return entity.StaticGroup{}, err
		}
	}

	return group, nil
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

	//TODO: add transaction
	err := g.staticGroupDao.UpdateStaticGroup(ct, group)
	if err != nil {
		return entity.StaticGroup{}, err
	}

	currentTeamIDs, err := g.teamGroupRelationDao.FindTeamIDsByGroupID(ct, group.ID)
	if err != nil {
		return entity.StaticGroup{}, err
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
			_, err := g.teamGroupRelationDao.CreateTeamGroupRelation(ct, entity.TeamGroupRelation{
				TeamID:  teamID,
				GroupID: group.ID,
			})
			if err != nil {
				return entity.StaticGroup{}, err
			}
		case delta.RemovedStatus:
			err := g.teamGroupRelationDao.DeleteTeamGroupRelation(ct, group.ID, teamID)
			if err != nil {
				return entity.StaticGroup{}, err
			}
		}
	}

	return group, nil
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

	//TODO: add transaction
	filterGroup, err := g.filterGroupDao.CreateFilterGroup(ct, filterGroup)
	if err != nil {
		return entity.FilterGroup{}, err
	}

	appGroupRelation := entity.AppGroupRelation{
		AppID:   appID,
		GroupID: filterGroup.ID,
		Type:    entity.AppGroupRelationTypeTeam,
	}
	_, err = g.appGroupRelationDao.CreateAppGroupRelation(ct, appGroupRelation)
	return filterGroup, err
}

func (g *Group) FindUsersByGroupID(ct context.Context, groupID uint64) ([]entity.User, *errs.Error) {
	var users []entity.User
	txCtx := TransactionsContext{
		logger:             g.logger,
		transactionFactory: g.transactionFactory,
		stateSyncer:        g.stateSyncer,
		ct:                 ct,
	}

	err := txCtx.withTransactions(true, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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
	txCtx := TransactionsContext{
		logger:             g.logger,
		transactionFactory: g.transactionFactory,
		stateSyncer:        g.stateSyncer,
		ct:                 ct,
	}

	err := txCtx.withTransactions(true, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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

	staticGroups, err := g.staticGroupDao.FindStaticGroupsByIDs(ct, groupIDs)
	if err != nil {
		return nil, err
	}

	filterGroups, err := g.filterGroupDao.FindFilterGroupsByIDs(ct, groupIDs)
	if err != nil {
		return nil, err
	}

	groupUnions := make([]entity.GroupUnion, 0)
	for _, staticGroup := range staticGroups {
		groupUnions = append(groupUnions, entity.GroupUnion{
			Type:        entity.GroupTypeStatic,
			StaticGroup: staticGroup,
		})
	}

	for _, filterGroup := range filterGroups {
		groupUnions = append(groupUnions, entity.GroupUnion{
			Type:        entity.GroupTypeFilter,
			FilterGroup: filterGroup,
		})
	}

	return groupUnions, nil
}

func NewGroup(
	logger telemetry.Logger,
	cloudClientRegistry *client.Registry,
	transactionFactory transaction.Factory,
	stateSyncer *realtime.StateSyncer,
	staticGroupDao dao.StaticGroup,
	filterGroupDao dao.FilterGroup,
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
		staticGroupDao:       staticGroupDao,
		filterGroupDao:       filterGroupDao,
		userGroupRelationDao: userGroupRelationDao,
		appGroupRelationDao:  appGroupRelationDao,
		teamGroupRelationDao: teamGroupRelationDao,
		userDao:              userDao,
		teamDao:              teamDao,
		appDao:               appDao,
	}
}
