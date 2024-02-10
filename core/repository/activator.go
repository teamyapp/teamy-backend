package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	daoEntity "github.com/teamyapp/teamy-backend/core/dao/entity"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Activator struct {
	activatorDao           dao.Activator
	timeRangeActivatorDao  dao.TimeRangeActivator
	maxViewersActivatorDao dao.MaxViewersActivator
	percentageActivatorDao dao.PercentageActivator
}

type CreatePartialActivatorInput struct {
	ID         uint64
	Type       entity.ActivatorType
	StartAt    *time.Time
	EndAt      *time.Time
	Percentage int
	MaxViewers int
}

func (a *Activator) FindActivatorByIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	activatorID uint64,
) (entity.ActivatorUnion, *errs.Error) {

	activator, err := a.activatorDao.FindActivatorByIDWithTx(ct, tx, activatorID)
	if err != nil {
		return entity.ActivatorUnion{}, err
	}

	return a.GetActivatorUnionFromBaseActivator(ct, tx, activator)
}

func (a *Activator) CreatePartialActivator(ct context.Context, tx *transaction.Transaction, input CreatePartialActivatorInput) *errs.Error {
	var err *errs.Error
	switch input.Type {
	case entity.ActivatorTypeStatic:
		break
	case entity.ActivatorTypeTimeRange:
		err = a.timeRangeActivatorDao.CreateTimeRangeActivator(ct, tx, input.ID, daoEntity.PartialTimeRangeActivator{
			StartAt: input.StartAt,
			EndAt:   input.EndAt,
		})

	case entity.ActivatorTypeMaxViewers:
		err = a.maxViewersActivatorDao.CreateMaxViewersActivator(ct, tx, input.ID, daoEntity.PartialMaxViewersActivator{
			MaxViewers: input.MaxViewers,
		})

	case entity.ActivatorTypePercentage:
		err = a.percentageActivatorDao.CreatePercentageActivator(ct, tx, input.ID, daoEntity.PartialPercentageActivator{
			Percentage: input.Percentage,
		})

	default:
		return errs.NewError(errs.Unknown, fmt.Sprintf("unknown activator type %s", input.Type))
	}

	return err
}

func (a *Activator) CreateStaticActivator(ct context.Context, tx *transaction.Transaction, staticActivator entity.StaticActivator) *errs.Error {
	return a.activatorDao.CreateActivator(ct, tx, staticActivator.Activator)
}

func (a *Activator) CreateTimeRangeActivator(ct context.Context, tx *transaction.Transaction, timeRangeActivator entity.TimeRangeActivator) *errs.Error {
	err := a.activatorDao.CreateActivator(ct, tx, timeRangeActivator.Activator)
	if err != nil {
		return err
	}

	return a.timeRangeActivatorDao.CreateTimeRangeActivator(ct, tx, timeRangeActivator.Activator.ID, daoEntity.PartialTimeRangeActivator{
		StartAt: timeRangeActivator.StartAt,
		EndAt:   timeRangeActivator.EndAt,
	})
}

func (a *Activator) UpdateStaticActivator(ct context.Context, tx *transaction.Transaction, staticActivator entity.StaticActivator) *errs.Error {
	return a.activatorDao.UpdateActivator(ct, tx, staticActivator.Activator)
}

func (a *Activator) UpdateTimeRangeActivator(ct context.Context, tx *transaction.Transaction, timeRangeActivator entity.TimeRangeActivator) *errs.Error {
	err := a.activatorDao.UpdateActivator(ct, tx, timeRangeActivator.Activator)
	if err != nil {
		return err
	}

	return a.timeRangeActivatorDao.UpdateTimeRangeActivator(ct, tx, timeRangeActivator.Activator.ID, daoEntity.PartialTimeRangeActivator{
		StartAt: timeRangeActivator.StartAt,
		EndAt:   timeRangeActivator.EndAt,
	})
}

func (a *Activator) CreateMaxViewersActivator(ct context.Context, tx *transaction.Transaction, maxViewersActivator entity.MaxViewersActivator) *errs.Error {
	err := a.activatorDao.CreateActivator(ct, tx, maxViewersActivator.Activator)
	if err != nil {
		return err
	}

	return a.maxViewersActivatorDao.CreateMaxViewersActivator(ct, tx, maxViewersActivator.Activator.ID, daoEntity.PartialMaxViewersActivator{
		MaxViewers: maxViewersActivator.MaxViewers,
	})
}

func (a *Activator) UpdateMaxViewersActivator(ct context.Context, tx *transaction.Transaction, maxViewersActivator entity.MaxViewersActivator) *errs.Error {
	err := a.activatorDao.UpdateActivator(ct, tx, maxViewersActivator.Activator)
	if err != nil {
		return err
	}

	return a.maxViewersActivatorDao.UpdateMaxViewersActivator(ct, tx, maxViewersActivator.Activator.ID, daoEntity.PartialMaxViewersActivator{
		MaxViewers: maxViewersActivator.MaxViewers,
	})
}

func (a *Activator) UpdateActivator(ct context.Context, tx *transaction.Transaction, activator entity.ActivatorUnion) {
}

func (a *Activator) CreatePercentageActivator(ct context.Context, tx *transaction.Transaction, percentageActivator entity.PercentageActivator) *errs.Error {
	err := a.activatorDao.CreateActivator(ct, tx, percentageActivator.Activator)
	if err != nil {
		return err
	}

	return a.percentageActivatorDao.CreatePercentageActivator(ct, tx, percentageActivator.Activator.ID, daoEntity.PartialPercentageActivator{
		Percentage: percentageActivator.Percentage,
	})
}

func (a *Activator) UpdatePercentageActivator(ct context.Context, tx *transaction.Transaction, percentageActivator entity.PercentageActivator) *errs.Error {
	err := a.activatorDao.UpdateActivator(ct, tx, percentageActivator.Activator)
	if err != nil {
		return err
	}

	return a.percentageActivatorDao.UpdatePercentageActivator(ct, tx, percentageActivator.Activator.ID, daoEntity.PartialPercentageActivator{
		Percentage: percentageActivator.Percentage,
	})
}

func (a *Activator) DeleteActivator(ct context.Context, tx *transaction.Transaction, activatorID uint64) (entity.ActivatorUnion, *errs.Error) {
	activator, err := a.activatorDao.FindActivatorByIDWithTx(ct, tx, activatorID)
	if err != nil {
		return entity.ActivatorUnion{}, err
	}

	err = a.activatorDao.DeleteActivator(ct, tx, activatorID)
	if err != nil {
		return entity.ActivatorUnion{}, err
	}

	activatorUnion, err := a.GetActivatorUnionFromBaseActivator(ct, tx, activator)
	if err != nil {
		return entity.ActivatorUnion{}, err
	}

	err = a.deletePartialActivator(ct, tx, activatorID, activator.Type)
	return activatorUnion, err
}

func (a *Activator) DeletePartialActivator(ct context.Context, tx *transaction.Transaction, activatorID uint64) *errs.Error {
	activator, err := a.activatorDao.FindActivatorByIDWithTx(ct, tx, activatorID)
	if err != nil {
		return err
	}

	return a.deletePartialActivator(ct, tx, activatorID, activator.Type)
}

func (a *Activator) deletePartialActivator(ct context.Context, tx *transaction.Transaction, activatorID uint64, activatorType entity.ActivatorType) *errs.Error {
	var err *errs.Error
	switch activatorType {
	case entity.ActivatorTypeStatic:
		break
	case entity.ActivatorTypeTimeRange:
		err = a.timeRangeActivatorDao.DeleteTimeRangeActivator(ct, tx, activatorID)
	case entity.ActivatorTypeMaxViewers:
		err = a.maxViewersActivatorDao.DeleteMaxViewersActivator(ct, tx, activatorID)
	case entity.ActivatorTypePercentage:
		err = a.percentageActivatorDao.DeletePercentageActivator(ct, tx, activatorID)
	default:
		err = errs.NewError(errs.Unknown, fmt.Sprintf("unknown activator type %s", activatorType))
	}

	return err
}

func (a *Activator) GetActivatorUnionFromBaseActivator(
	ct context.Context,
	tx *transaction.Transaction,
	activator entity.Activator,
) (entity.ActivatorUnion, *errs.Error) {
	var err *errs.Error
	activatorUnion := entity.ActivatorUnion{
		Type: activator.Type,
	}

	activatorID := activator.ID
	switch activator.Type {
	case entity.ActivatorTypeStatic:
		activatorUnion.StaticActivator = entity.StaticActivator{
			Activator: activator,
		}
	case entity.ActivatorTypeTimeRange:
		partialTimeRangeActivator, err := a.timeRangeActivatorDao.FindTimeRangeActivatorByIDWithTx(ct, tx, activatorID)
		if err != nil {
			return entity.ActivatorUnion{}, err
		}

		activatorUnion.TimeRangeActivator = entity.TimeRangeActivator{
			Activator: activator,
			StartAt:   partialTimeRangeActivator.StartAt,
			EndAt:     partialTimeRangeActivator.EndAt,
		}
	case entity.ActivatorTypeMaxViewers:
		partialMaxViewersActivator, err := a.maxViewersActivatorDao.FindMaxViewersActivatorByIDWithTx(ct, tx, activatorID)
		if err != nil {
			return entity.ActivatorUnion{}, err
		}

		activatorUnion.MaxViewersActivator = entity.MaxViewersActivator{
			Activator:  activator,
			MaxViewers: partialMaxViewersActivator.MaxViewers,
		}
	case entity.ActivatorTypePercentage:
		partialPercentageActivator, err := a.percentageActivatorDao.FindPercentageActivatorByIDWithTx(ct, tx, activatorID)
		if err != nil {
			return entity.ActivatorUnion{}, err
		}

		activatorUnion.PercentageActivator = entity.PercentageActivator{
			Activator:  activator,
			Percentage: partialPercentageActivator.Percentage,
		}
	default:
		err = errs.NewError(errs.Unknown, fmt.Sprintf("unknown activator type %s", activator.Type))
	}

	return activatorUnion, err
}

func NewActivator(
	activatorDao dao.Activator,
	timeRangeActivatorDao dao.TimeRangeActivator,
	maxViewersActivatorDao dao.MaxViewersActivator,
	percentageActivatorDao dao.PercentageActivator,
) *Activator {
	return &Activator{
		activatorDao:           activatorDao,
		timeRangeActivatorDao:  timeRangeActivatorDao,
		maxViewersActivatorDao: maxViewersActivatorDao,
		percentageActivatorDao: percentageActivatorDao,
	}
}
