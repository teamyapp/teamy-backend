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

func (a *Activator) FindActivatorByID(ct context.Context, activatorID uint64) (entity.ActivatorUnion, *errs.Error) {
	activatorType, err := a.activatorTypeRelationDao.FindActivatorTypeByID(ct, activatorID)
	if err != nil {
		return entity.ActivatorUnion{}, err
	}

	activatorUnion := entity.ActivatorUnion{
		Type: activatorType,
	}
	switch activatorType {
	case entity.ActivatorTypeTimeRange:
		activatorUnion.TimeRangeActivator, err = a.timeRangeActivatorDao.FindTimeRangeActivatorByID(ct, activatorID)
	case entity.ActivatorTypeMaxViewers:
		activatorUnion.MaxViewersActivator, err = a.maxViewersActivatorDao.FindMaxViewersActivatorByID(ct, activatorID)
	case entity.ActivatorTypePercentage:
		activatorUnion.PercentageActivator, err = a.percentageActivatorDao.FindPercentageActivatorByID(ct, activatorID)
	default:
		err = errs.NewError(errs.Unknown, fmt.Sprintf("unknown activator type %s", activatorType))
	}

	return activatorUnion, err
}

func (a *Activator) CreateTimeRangeActivator(ct context.Context, timeRangeActivator entity.TimeRangeActivator) (entity.TimeRangeActivator, *errs.Error) {
	_, err := a.activatorDao.CreateActivator(ct, timeRangeActivator.Activator)
	if err != nil {
		return entity.TimeRangeActivator{}, err
	}

	activatorTypeRelation := entity.ActivatorTypeRelation{
		ActivatorID:   timeRangeActivator.Activator.ID,
		ActivatorType: entity.ActivatorTypeTimeRange,
	}
	_, err = a.activatorTypeRelationDao.CreateActivatorTypeRelation(ct, activatorTypeRelation)
	if err != nil {
		return entity.TimeRangeActivator{}, err
	}
	
	return a.timeRangeActivatorDao.CreateTimeRangeActivator(ct, timeRangeActivator)
}

func (a *Activator) CreateMaxViewersActivator(ct context.Context, maxViewersActivator entity.MaxViewersActivator) (entity.MaxViewersActivator, *errs.Error) {
	_, err := a.activatorDao.CreateActivator(ct, maxViewersActivator.Activator)
	if err != nil {
		return entity.MaxViewersActivator{}, err
	}

	activatorTypeRelation := entity.ActivatorTypeRelation{
		ActivatorID:   maxViewersActivator.Activator.ID,
		ActivatorType: entity.ActivatorTypeMaxViewers,
	}
	_, err = a.activatorTypeRelationDao.CreateActivatorTypeRelation(ct, activatorTypeRelation)
	if err != nil {
		return entity.MaxViewersActivator{}, err
	}

	return a.maxViewersActivatorDao.CreateMaxViewersActivator(ct, maxViewersActivator)
}

func (a *Activator) CreatePercentageActivator(ct context.Context, percentageActivator entity.PercentageActivator) (entity.PercentageActivator, *errs.Error) {
	_, err := a.activatorDao.CreateActivator(ct, percentageActivator.Activator)
	if err != nil {
		return entity.PercentageActivator{}, err
	}

	activatorTypeRelation := entity.ActivatorTypeRelation{
		ActivatorID:   percentageActivator.Activator.ID,
		ActivatorType: entity.ActivatorTypePercentage,
	}
	_, err = a.activatorTypeRelationDao.CreateActivatorTypeRelation(ct, activatorTypeRelation)
	if err != nil {
		return entity.PercentageActivator{}, err
	}

	return a.percentageActivatorDao.CreatePercentageActivator(ct, percentageActivator)
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
