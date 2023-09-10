package repository

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Activator struct {
	logger                   telemetry.Logger
	activatorDao             dao.Activator
	timeRangeActivatorDao    dao.TimeRangeActivator
	maxViewersActivatorDao   dao.MaxViewersActivator
	percentageActivatorDao   dao.PercentageActivator
	activatorTypeRelationDao dao.ActivatorTypeRelation
}

func (a *Activator) FindActivatorByID(ct context.Context, ActivatorID uint64) (entity.ActivatorUnion, *errs.Error) {
	activatorType, err := a.activatorTypeRelationDao.FindActivatorTypeByID(ct, ActivatorID)
	if err != nil {
		return entity.ActivatorUnion{}, err
	}

	var timeRangeActivator entity.TimeRangeActivator = entity.TimeRangeActivator{}
	var maxViewersActivator entity.MaxViewersActivator = entity.MaxViewersActivator{}
	var percentageActivator entity.PercentageActivator = entity.PercentageActivator{}
	switch activatorType {
	case entity.ActivatorTypeTimeRange:
		timeRangeActivator, err = a.timeRangeActivatorDao.FindTimeRangeActivatorByID(ct, ActivatorID)
	case entity.ActivatorTypeMaxViewers:
		maxViewersActivator, err = a.maxViewersActivatorDao.FindMaxViewersActivatorByID(ct, ActivatorID)
	case entity.ActivatorTypePercentage:
		percentageActivator, err = a.percentageActivatorDao.FindPercentageActivatorByID(ct, ActivatorID)
	default:
		err = errs.NewError(errs.Unknown, fmt.Sprintf("unknown activator type %s", activatorType))
	}

	if err != nil {
		return entity.ActivatorUnion{}, err
	}
	return entity.ActivatorUnion{
		Type:                activatorType,
		TimeRangeActivator:  timeRangeActivator,
		MaxViewersActivator: maxViewersActivator,
		PercentageActivator: percentageActivator,
	}, nil
}

func (a *Activator) CreateTimeRangeActivator(ct context.Context, timeRangeActivator entity.TimeRangeActivator) (entity.TimeRangeActivator, *errs.Error) {
	activatorTypeRelation := entity.ActivatorTypeRelation{
		ActivatorID:   timeRangeActivator.Activator.ID,
		ActivatorType: entity.ActivatorTypeTimeRange,
	}
	_, err := a.activatorTypeRelationDao.CreateActivatorTypeRelation(ct, activatorTypeRelation)
	if err != nil {
		return entity.TimeRangeActivator{}, err
	}

	_, err = a.activatorDao.CreateActivator(ct, timeRangeActivator.Activator)
	if err != nil {
		return entity.TimeRangeActivator{}, err
	}

	return a.timeRangeActivatorDao.CreateTimeRangeActivator(ct, timeRangeActivator)
}

func (a *Activator) CreateMaxViewersActivator(ct context.Context, maxViewersActivator entity.MaxViewersActivator) (entity.MaxViewersActivator, *errs.Error) {
	activatorTypeRelation := entity.ActivatorTypeRelation{
		ActivatorID:   maxViewersActivator.Activator.ID,
		ActivatorType: entity.ActivatorTypeMaxViewers,
	}
	_, err := a.activatorTypeRelationDao.CreateActivatorTypeRelation(ct, activatorTypeRelation)
	if err != nil {
		return entity.MaxViewersActivator{}, err
	}

	_, err = a.activatorDao.CreateActivator(ct, maxViewersActivator.Activator)
	if err != nil {
		return entity.MaxViewersActivator{}, err
	}

	return a.maxViewersActivatorDao.CreateMaxViewersActivator(ct, maxViewersActivator)
}

func (a *Activator) CreatePercentageActivator(ct context.Context, percentageActivator entity.PercentageActivator) (entity.PercentageActivator, *errs.Error) {
	activatorTypeRelation := entity.ActivatorTypeRelation{
		ActivatorID:   percentageActivator.Activator.ID,
		ActivatorType: entity.ActivatorTypePercentage,
	}
	_, err := a.activatorTypeRelationDao.CreateActivatorTypeRelation(ct, activatorTypeRelation)
	if err != nil {
		return entity.PercentageActivator{}, err
	}

	_, err = a.activatorDao.CreateActivator(ct, percentageActivator.Activator)
	if err != nil {
		return entity.PercentageActivator{}, err
	}

	return a.percentageActivatorDao.CreatePercentageActivator(ct, percentageActivator)
}

func (a *Activator) FindActivatorTypeByID(ct context.Context, activatorID uint64) (entity.ActivatorType, *errs.Error) {
	return a.activatorTypeRelationDao.FindActivatorTypeByID(ct, activatorID)
}

func (a *Activator) FindTimeRangeActivatorByID(ct context.Context, activatorID uint64) (entity.TimeRangeActivator, *errs.Error) {
	return a.timeRangeActivatorDao.FindTimeRangeActivatorByID(ct, activatorID)
}

func (a *Activator) FindMaxViewersActivatorByID(ct context.Context, activatorID uint64) (entity.MaxViewersActivator, *errs.Error) {
	return a.maxViewersActivatorDao.FindMaxViewersActivatorByID(ct, activatorID)
}

func (a *Activator) FindPercentageActivatorByID(ct context.Context, activatorID uint64) (entity.PercentageActivator, *errs.Error) {
	return a.percentageActivatorDao.FindPercentageActivatorByID(ct, activatorID)
}

func NewActivator(
	logger telemetry.Logger,
	timeRangeActivatorDao dao.TimeRangeActivator,
	maxViewersActivatorDao dao.MaxViewersActivator,
	percentageActivatorDao dao.PercentageActivator,
	activatorTypeRelationDao dao.ActivatorTypeRelation,
) *Activator {
	return &Activator{
		logger:                   logger,
		timeRangeActivatorDao:    timeRangeActivatorDao,
		maxViewersActivatorDao:   maxViewersActivatorDao,
		percentageActivatorDao:   percentageActivatorDao,
		activatorTypeRelationDao: activatorTypeRelationDao,
	}
}
