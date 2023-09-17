package repository

import (
	"context"
	"fmt"

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

func (g *Group) CreateStaticGroup(ct context.Context, tx *transaction.Transaction, staticGroup entity.StaticGroup) (entity.StaticGroup, *errs.Error) {
	err := g.groupDao.CreateGroup(ct, tx, staticGroup.Group)
	if err != nil {
		return entity.StaticGroup{}, err
	}

	return staticGroup, nil
}

func (g *Group) UpdateStaticGroup(ct context.Context, tx *transaction.Transaction, staticGroup entity.StaticGroup) *errs.Error {
	return g.groupDao.UpdateGroup(ct, tx, staticGroup.Group)
}

func (g *Group) CreateFilterGroup(ct context.Context, tx *transaction.Transaction, filterGroup entity.FilterGroup) (entity.FilterGroup, *errs.Error) {
	err := g.groupDao.CreateGroup(ct, tx, filterGroup.Group)
	if err != nil {
		return entity.FilterGroup{}, err
	}

	err = g.filterGroupDao.CreateFilterGroup(ct, tx, filterGroup)
	if err != nil {
		return entity.FilterGroup{}, err
	}

	return filterGroup, nil
}

func (g *Group) UpdateFilterGroup(ct context.Context, tx *transaction.Transaction, filterGroup entity.FilterGroup) *errs.Error {
	err := g.groupDao.UpdateGroup(ct, tx, filterGroup.Group)
	if err != nil {
		return err
	}

	return g.filterGroupDao.UpdateFilterGroup(ct, tx, filterGroup)
}

func (g *Group) FindGroupsByIDs(ct context.Context, groupIDs []uint64) ([]entity.GroupUnion, *errs.Error) {
	groups, err := g.groupDao.FindGroupsByIDs(ct, groupIDs)
	if err != nil {
		return nil, err
	}

	groupUnions := make([]entity.GroupUnion, 0)
	for _, group := range groups {
		switch group.Type {
		case entity.GroupTypeStatic:
			groupUnions = append(groupUnions, entity.GroupUnion{
				Type: entity.GroupTypeStatic,
				StaticGroup: entity.StaticGroup{
					Group: group,
				},
			})
		case entity.GroupTypeFilter:
			filterGroup, err := g.filterGroupDao.FindFilterGroupByID(ct, group.ID)
			if err != nil {
				return nil, err
			}

			groupUnions = append(groupUnions, entity.GroupUnion{
				Type: entity.GroupTypeFilter,
				FilterGroup: entity.FilterGroup{
					Group:  group,
					Filter: filterGroup.Filter,
					Count:  filterGroup.Count,
				},
			})
		default:
			return nil, errs.NewError(errs.Unknown, fmt.Sprintf("unknown group type %s", group.Type))
		}
	}

	return groupUnions, nil
}

func NewGroup(logger telemetry.Logger, groupDao dao.Group, filterGroupDao dao.FilterGroup) *Group {
	return &Group{
		logger:         logger,
		groupDao:       groupDao,
		filterGroupDao: filterGroupDao,
	}
}
