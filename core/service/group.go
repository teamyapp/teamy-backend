package service

import (
	"context"
	"time"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type Group struct {
	logger               telemetry.Logger
	cloudClientRegistry  *client.Registry
	transactionFactory   transaction.Factory
	stateSyncer          *realtime.StateSyncer
	groupDao             dao.Group
	filterGroupDao       dao.FilterGroup
	userGroupRelationDao dao.UserGroupRelation
	appGroupRelationDao  dao.AppGroupRelation
	teamGroupRelationDao dao.TeamGroupRelation
	userDao              dao.User
	teamDao              dao.Team
	appDao               dao.App
}

func (g *Group) CreateStaticUserGroup(ct context.Context, appID uint64, name string, userIDs []uint64) (entity.Group, *errs.Error) {
	genGroupIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "groupID"}
	genGroupIDRes, rpcErr := g.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genGroupIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.Group{}, internalErr
	}

	now := time.Now().UTC()
	group := entity.Group{
		ID:        genGroupIDRes.UniqueNumber,
		Name:      name,
		Type:      entity.GroupTypeStatic,
		CreatedAt: now,
	}

	group, err := g.groupDao.CreateGroup(ct, group)
	if err != nil {
		return entity.Group{}, err
	}

	appGroupRelation := entity.AppGroupRelation{
		AppID:   appID,
		GroupID: group.ID,
		Type:    entity.AppGroupRelationTypeUser,
	}

	_, err = g.appGroupRelationDao.CreateAppGroupRelation(ct, appGroupRelation)
	if err != nil {
		return entity.Group{}, err
	}

	for _, userID := range userIDs {
		_, err := g.userGroupRelationDao.CreateUserGroupRelation(ct, entity.UserGroupRelation{
			UserID:  userID,
			GroupID: group.ID,
		})
		if err != nil {
			return entity.Group{}, err
		}
	}

	return group, nil
}

func (g *Group) UpdateStaticUserGroup(ct context.Context, appID uint64, groupID uint64, name string, userIDs []uint64) (entity.Group, *errs.Error) {
	currentAppID, err := g.appGroupRelationDao.FindAppIDByGroupID(ct, groupID)
	if err != nil {
		return entity.Group{}, err
	}

	if currentAppID != appID {
		return entity.Group{}, errs.NewError(
			errs.InvalidArgument,
			"appID is not matched",
		)
	}

	now := time.Now().UTC()
	group := entity.Group{
		Name:      name,
		Type:      entity.GroupTypeStatic,
		CreatedAt: now,
	}

	err = g.groupDao.UpdateGroup(ct, groupID, group)
	if err != nil {
		return entity.Group{}, err
	}

	currentUserIDs, err := g.userGroupRelationDao.FindUserIDsByGroupID(ct, group.ID)
	if err != nil {
		return entity.Group{}, err
	}

	currentUserIDsSet := map[uint64]bool{}
	for _, userID := range currentUserIDs {
		currentUserIDsSet[userID] = true
	}

	for _, userID := range userIDs {
		if _, ok := currentUserIDsSet[userID]; ok {
			continue
		}

		_, err := g.userGroupRelationDao.CreateUserGroupRelation(ct, entity.UserGroupRelation{
			UserID:  userID,
			GroupID: group.ID,
		})

		if err != nil {
			return entity.Group{}, err
		}
	}

	teamIDsSet := map[uint64]bool{}
	for _, userID := range userIDs {
		teamIDsSet[userID] = true
	}

	for _, userID := range currentUserIDs {
		if _, ok := teamIDsSet[userID]; ok {
			continue
		}

		err := g.userGroupRelationDao.DeleteUserGroupRelation(ct, group.ID, userID)
		if err != nil {
			return entity.Group{}, err
		}
	}

	return group, nil
}

func (g *Group) CreateFilterUserGroup(ct context.Context, appID uint64, name string, filter string) (entity.FilterGroup, *errs.Error) {
	genGroupIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "groupID"}
	genGroupIDRes, rpcErr := g.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genGroupIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.FilterGroup{}, internalErr
	}

	now := time.Now().UTC()
	group := entity.Group{
		ID:        genGroupIDRes.UniqueNumber,
		Name:      name,
		Type:      entity.GroupTypeFilter,
		CreatedAt: now,
	}
	filterGroup := entity.FilterGroup{
		Group:  group,
		Filter: filter,
		Count:  0,
	}

	filterGroup, err := g.filterGroupDao.CreateFilterGroup(ct, filterGroup)
	if err != nil {
		return entity.FilterGroup{}, err
	}

	appGroupRelation := entity.AppGroupRelation{
		AppID:   appID,
		GroupID: group.ID,
		Type:    entity.AppGroupRelationTypeUser,
	}

	_, err = g.appGroupRelationDao.CreateAppGroupRelation(ct, appGroupRelation)
	if err != nil {
		return entity.FilterGroup{}, err
	}

	return filterGroup, nil
}

func (g *Group) UpdateFilterUserGroup(ct context.Context, appID uint64, groupID uint64, name string, filter string) (entity.FilterGroup, *errs.Error) {
	currentAppID, err := g.appGroupRelationDao.FindAppIDByGroupID(ct, groupID)
	if err != nil {
		return entity.FilterGroup{}, err
	}

	if currentAppID != appID {
		return entity.FilterGroup{}, errs.NewError(
			errs.InvalidArgument,
			"appID is not matched",
		)
	}

	now := time.Now().UTC()
	filterGroup := entity.FilterGroup{
		Group: entity.Group{
			Name:      name,
			Type:      entity.GroupTypeFilter,
			UpdatedAt: &now,
		},
		Filter: filter,
		Count:  0,
	}

	err = g.filterGroupDao.UpdateFilterGroup(ct, groupID, filterGroup)
	if err != nil {
		return entity.FilterGroup{}, err
	}

	return filterGroup, nil
}

func (g *Group) CreateStaticTeamGroup(ct context.Context, appID uint64, name string, teamIDs []uint64) (entity.Group, *errs.Error) {
	genGroupIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "groupID"}
	genGroupIDRes, rpcErr := g.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genGroupIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.Group{}, internalErr
	}

	now := time.Now().UTC()
	group := entity.Group{
		ID:        genGroupIDRes.UniqueNumber,
		Name:      name,
		Type:      entity.GroupTypeStatic,
		CreatedAt: now,
	}

	group, err := g.groupDao.CreateGroup(ct, group)
	if err != nil {
		return entity.Group{}, err
	}

	appGroupRelation := entity.AppGroupRelation{
		AppID:   appID,
		GroupID: group.ID,
		Type:    entity.AppGroupRelationTypeTeam,
	}

	_, err = g.appGroupRelationDao.CreateAppGroupRelation(ct, appGroupRelation)
	if err != nil {
		return entity.Group{}, err
	}

	for _, teamID := range teamIDs {
		_, err := g.teamGroupRelationDao.CreateTeamGroupRelation(ct, entity.TeamGroupRelation{
			TeamID:  teamID,
			GroupID: group.ID,
		})
		if err != nil {
			return entity.Group{}, err
		}
	}

	return group, nil
}

func (g *Group) UpdateStaticTeamGroup(ct context.Context, appID, groupID uint64, name string, teamIDs []uint64) (entity.Group, *errs.Error) {
	currentAppID, err := g.appGroupRelationDao.FindAppIDByGroupID(ct, groupID)
	if err != nil {
		return entity.Group{}, err
	}

	if currentAppID != appID {
		return entity.Group{}, errs.NewError(
			errs.InvalidArgument,
			"appID is not matched",
		)
	}

	now := time.Now().UTC()
	group := entity.Group{
		Name:      name,
		Type:      entity.GroupTypeStatic,
		CreatedAt: now,
	}

	err = g.groupDao.UpdateGroup(ct, groupID, group)
	if err != nil {
		return entity.Group{}, err
	}

	currentTeamIDs, err := g.teamGroupRelationDao.FindTeamIDsByGroupID(ct, group.ID)
	if err != nil {
		return entity.Group{}, err
	}

	currentTeamIDsSet := map[uint64]bool{}
	for _, teamID := range currentTeamIDs {
		currentTeamIDsSet[teamID] = true
	}

	for _, teamID := range teamIDs {
		if _, ok := currentTeamIDsSet[teamID]; ok {
			continue
		}

		_, err := g.teamGroupRelationDao.CreateTeamGroupRelation(ct, entity.TeamGroupRelation{
			TeamID:  teamID,
			GroupID: group.ID,
		})

		if err != nil {
			return entity.Group{}, err
		}
	}

	teamIDsSet := map[uint64]bool{}
	for _, teamID := range teamIDs {
		teamIDsSet[teamID] = true
	}

	for _, teamID := range currentTeamIDs {
		if _, ok := teamIDsSet[teamID]; ok {
			continue
		}

		err := g.teamGroupRelationDao.DeleteTeamGroupRelation(ct, group.ID, teamID)
		if err != nil {
			return entity.Group{}, err
		}
	}

	return group, nil
}

func (g *Group) UpdateFilterTeamGroup(
	ct context.Context, appID, groupID uint64, name string, filter string,
) (entity.FilterGroup, *errs.Error) {
	currentAppID, err := g.appGroupRelationDao.FindAppIDByGroupID(ct, groupID)
	if err != nil {
		return entity.FilterGroup{}, err
	}

	if currentAppID != appID {
		return entity.FilterGroup{}, errs.NewError(
			errs.InvalidArgument,
			"appID is not matched",
		)
	}

	now := time.Now().UTC()
	filterGroup := entity.FilterGroup{
		Group: entity.Group{
			Name:      name,
			Type:      entity.GroupTypeFilter,
			UpdatedAt: &now,
		},
		Filter: filter,
		Count:  0,
	}

	err = g.filterGroupDao.UpdateFilterGroup(ct, groupID, filterGroup)
	if err != nil {
		return entity.FilterGroup{}, err
	}

	return filterGroup, nil
}

func (g *Group) CreateFilterTeamGroup(ct context.Context, appID uint64, name string, filter string) (entity.FilterGroup, *errs.Error) {
	genGroupIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "groupID"}
	genGroupIDRes, rpcErr := g.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genGroupIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.FilterGroup{}, internalErr
	}

	now := time.Now().UTC()

	group := entity.Group{
		ID:        genGroupIDRes.UniqueNumber,
		Name:      name,
		Type:      entity.GroupTypeFilter,
		CreatedAt: now,
	}
	filterGroup := entity.FilterGroup{
		Group:  group,
		Filter: filter,
		Count:  0,
	}

	filterGroup, err := g.filterGroupDao.CreateFilterGroup(ct, filterGroup)
	if err != nil {
		return entity.FilterGroup{}, err
	}

	appGroupRelation := entity.AppGroupRelation{
		AppID:   appID,
		GroupID: group.ID,
		Type:    entity.AppGroupRelationTypeTeam,
	}

	_, err = g.appGroupRelationDao.CreateAppGroupRelation(ct, appGroupRelation)
	if err != nil {
		return entity.FilterGroup{}, err
	}

	return filterGroup, nil
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

	if err != nil {
		return nil, err
	}

	return users, nil
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

	if err != nil {
		return nil, err
	}

	return teams, nil
}

func (g *Group) FindAppByGroupID(ct context.Context, groupID uint64) (entity.App, *errs.Error) {
	appID, err := g.appGroupRelationDao.FindAppIDByGroupID(ct, groupID)
	if err != nil {
		return entity.App{}, err
	}

	return g.appDao.FindAppByID(ct, appID)
}

func (g *Group) FindFilterUserGroupsByAppID(ct context.Context, appID uint64) ([]entity.FilterGroup, *errs.Error) {
	groupIDs, err := g.appGroupRelationDao.FindGroupIDsByAppIDAndType(ct, appID, entity.AppGroupRelationTypeUser)
	if err != nil {
		return nil, err
	}

	return g.filterGroupDao.FindFilterGroupsByIDs(ct, groupIDs)
}

func (g *Group) FindStaticUserGroupsByAppID(ct context.Context, appID uint64) ([]entity.Group, *errs.Error) {
	groupIDs, err := g.appGroupRelationDao.FindGroupIDsByAppIDAndType(ct, appID, entity.AppGroupRelationTypeUser)
	if err != nil {
		return nil, err
	}

	return g.groupDao.FindGroupsByIDs(ct, groupIDs)
}

func (g *Group) FindStaticTeamGroupsByAppID(ct context.Context, appID uint64) ([]entity.Group, *errs.Error) {
	groupIDs, err := g.appGroupRelationDao.FindGroupIDsByAppIDAndType(ct, appID, entity.AppGroupRelationTypeTeam)
	if err != nil {
		return nil, err
	}

	return g.groupDao.FindGroupsByIDs(ct, groupIDs)
}

func (g *Group) FindFilterTeamGroupsByAppID(ct context.Context, appID uint64) ([]entity.FilterGroup, *errs.Error) {
	groupIDs, err := g.appGroupRelationDao.FindGroupIDsByAppIDAndType(ct, appID, entity.AppGroupRelationTypeTeam)
	if err != nil {
		return nil, err
	}

	return g.filterGroupDao.FindFilterGroupsByIDs(ct, groupIDs)
}

func NewGroup(
	logger telemetry.Logger,
	cloudClientRegistry *client.Registry,
	transactionFactory transaction.Factory,
	stateSyncer *realtime.StateSyncer,
	groupDao dao.Group,
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
		groupDao:             groupDao,
		filterGroupDao:       filterGroupDao,
		userGroupRelationDao: userGroupRelationDao,
		appGroupRelationDao:  appGroupRelationDao,
		teamGroupRelationDao: teamGroupRelationDao,
		userDao:              userDao,
		teamDao:              teamDao,
		appDao:               appDao,
	}
}
