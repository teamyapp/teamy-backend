package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Group struct {
	logger         telemetry.Logger
	groupDao       dao.Group
	filterGroupDao dao.FilterGroup
}

type CreatePartialGroupInput struct {
	ID     uint64
	Type   entity.GroupType
	Filter string
}

func (g *Group) CreateStaticGroup(ct context.Context, tx *transaction.Transaction, staticGroup entity.StaticGroup) *errs.Error {
	return g.groupDao.CreateGroup(ct, tx, staticGroup.Group)
}

func (g *Group) UpdateStaticGroup(ct context.Context, tx *transaction.Transaction, staticGroup entity.StaticGroup) *errs.Error {
	return g.groupDao.UpdateGroup(ct, tx, staticGroup.Group)
}

func (g *Group) CreatePartialGroup(ct context.Context, tx *transaction.Transaction, input CreatePartialGroupInput) *errs.Error {
	group := entity.Group{
		ID: input.ID,
	}

	switch input.Type {
	case entity.GroupTypeStatic:
		return nil
	case entity.GroupTypeFilter:
		filterGroup := entity.FilterGroup{
			Group:  group,
			Filter: input.Filter,
		}
		return g.filterGroupDao.CreateFilterGroup(ct, tx, filterGroup)
	default:
		return errs.NewError(errs.Unknown, fmt.Sprintf("unknown group type %s", input.Type))
	}
}

func (g *Group) CreateFilterGroup(ct context.Context, tx *transaction.Transaction, filterGroup entity.FilterGroup) *errs.Error {
	err := g.groupDao.CreateGroup(ct, tx, filterGroup.Group)
	if err != nil {
		return err
	}

	return g.filterGroupDao.CreateFilterGroup(ct, tx, filterGroup)
}

func (g *Group) UpdateFilterGroup(ct context.Context, tx *transaction.Transaction, filterGroup entity.FilterGroup) *errs.Error {
	err := g.groupDao.UpdateGroup(ct, tx, filterGroup.Group)
	if err != nil {
		return err
	}

	return g.filterGroupDao.UpdateFilterGroup(ct, tx, filterGroup)
}

func (g *Group) UpdateMaxRolloutIndexByGroupID(ct context.Context, tx *transaction.Transaction, groupID uint64, step int) (int, *errs.Error) {
	group, err := g.groupDao.FindGroupByIDWithTx(ct, tx, groupID)
	if err != nil {
		return 0, err
	}

	group.MaxRolloutIndex += step
	now := time.Now().UTC()
	group.UpdatedAt = &now
	err = g.groupDao.UpdateGroup(ct, tx, group)
	if err != nil {
		return 0, err
	}

	return group.MaxRolloutIndex, nil
}

func (g *Group) FindGroupByIDWithTx(ct context.Context, tx *transaction.Transaction, groupID uint64) (entity.GroupUnion, *errs.Error) {
	group, err := g.groupDao.FindGroupByIDWithTx(ct, tx, groupID)
	if err != nil {
		return entity.GroupUnion{}, err
	}

	return g.GetGroupUnionFromBaseGroup(ct, tx, group)
}

func (g *Group) FindGroupsByIDsWithTx(ct context.Context, tx *transaction.Transaction, groupIDs []uint64) ([]entity.GroupUnion, *errs.Error) {
	groups, err := g.groupDao.FindGroupsByIDsWithTx(ct, tx, groupIDs)
	if err != nil {
		return nil, err
	}

	groupUnions := make([]entity.GroupUnion, 0)
	for _, group := range groups {
		groupUnion, err := g.GetGroupUnionFromBaseGroup(ct, tx, group)
		if err != nil {
			return nil, err
		}

		groupUnions = append(groupUnions, groupUnion)
	}

	return groupUnions, nil
}

func (g *Group) DeleteGroup(ct context.Context, tx *transaction.Transaction, groupID uint64) (entity.GroupUnion, *errs.Error) {
	group, err := g.groupDao.FindGroupByIDWithTx(ct, tx, groupID)
	if err != nil {
		return entity.GroupUnion{}, err
	}

	err = g.groupDao.DeleteGroup(ct, tx, groupID)
	if err != nil {
		return entity.GroupUnion{}, err
	}

	groupUnion, err := g.GetGroupUnionFromBaseGroup(ct, tx, group)
	if err != nil {
		return entity.GroupUnion{}, err
	}

	err = g.deletePartialGroup(ct, tx, groupID, group.Type)
	return groupUnion, err
}

func (g *Group) DeletePartialGroup(ct context.Context, tx *transaction.Transaction, groupID uint64) *errs.Error {
	group, err := g.groupDao.FindGroupByIDWithTx(ct, tx, groupID)
	if err != nil {
		return err
	}

	return g.deletePartialGroup(ct, tx, groupID, group.Type)
}

func (g *Group) deletePartialGroup(ct context.Context, tx *transaction.Transaction, groupID uint64, groupType entity.GroupType) *errs.Error {
	switch groupType {
	case entity.GroupTypeStatic:
		return nil
	case entity.GroupTypeFilter:
		return g.filterGroupDao.DeleteFilterGroup(ct, tx, groupID)
	default:
		return errs.NewError(errs.Unknown, fmt.Sprintf("unknown group type %s", groupType))
	}
}

func (g *Group) GetGroupUnionFromBaseGroup(ct context.Context, tx *transaction.Transaction, group entity.Group) (entity.GroupUnion, *errs.Error) {
	switch group.Type {
	case entity.GroupTypeStatic:
		return entity.GroupUnion{
			Type:       entity.GroupTypeStatic,
			MemberType: group.MemberType,
			StaticGroup: entity.StaticGroup{
				Group: group,
			},
		}, nil
	case entity.GroupTypeFilter:
		filterGroup, err := g.filterGroupDao.FindFilterGroupByIDWithTx(ct, tx, group.ID)
		if err != nil {
			return entity.GroupUnion{}, err
		}

		return entity.GroupUnion{
			Type:       entity.GroupTypeFilter,
			MemberType: group.MemberType,
			FilterGroup: entity.FilterGroup{
				Group:  group,
				Filter: filterGroup.Filter,
			},
		}, nil
	default:
		return entity.GroupUnion{}, errs.NewError(errs.Unknown, fmt.Sprintf("unknown group type %s", group.Type))
	}
}

func NewGroup(logger telemetry.Logger, groupDao dao.Group, filterGroupDao dao.FilterGroup) *Group {
	return &Group{
		logger:         logger,
		groupDao:       groupDao,
		filterGroupDao: filterGroupDao,
	}
}
