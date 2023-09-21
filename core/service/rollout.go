package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/randgen"
	"github.com/teamyapp/cloud/libs/rollout"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudTransaction "github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/repository"
	"github.com/teamyapp/teamy-backend/core/store"
	"github.com/teamyapp/teamy-backend/core/transaction"
)

type CreateRolloutInput struct {
	VersionSelectorID uint64
	ActivatorID       uint64
	IsEnabled         bool
}

type UpdateRolloutInput struct {
	VersionSelectorID uint64
	ActivatorID       uint64
	IsEnabled         bool
}

type Rollout struct {
	logger                    telemetry.Logger
	transactionFactory        cloudTransaction.Factory
	stateSyncer               *realtime.StateSyncer
	appGroupRelationDao       dao.AppGroupRelation
	rolloutDao                dao.Rollout
	rolloutViewerDao          dao.RolloutViewer
	groupRolloutRelationDao   dao.GroupRolloutRelation
	appRolloutRelationDao     dao.AppRolloutRelation
	appVersionDao             dao.AppVersion
	versionSelectorRepository *repository.VersionSelector
	activatorRepository       *repository.Activator
}

func (r *Rollout) FindUserRolloutsByAppID(ct context.Context, appID uint64) ([]entity.Rollout, *errs.Error) {
	rolloutIDs, err := r.appRolloutRelationDao.FindRolloutIDsByAppIDAndRelationType(ct, appID, entity.AppRolloutRelationTypeUser)
	if err != nil {
		return nil, err
	}

	return r.rolloutDao.FindRolloutsByIDs(ct, rolloutIDs)
}

func (r *Rollout) FindTeamRolloutsByAppID(ct context.Context, appID uint64) ([]entity.Rollout, *errs.Error) {
	rolloutIDs, err := r.appRolloutRelationDao.FindRolloutIDsByAppIDAndRelationType(ct, appID, entity.AppRolloutRelationTypeTeam)
	if err != nil {
		return nil, err
	}

	return r.rolloutDao.FindRolloutsByIDs(ct, rolloutIDs)
}

func (r *Rollout) CreateAppRollout(
	ct context.Context,
	appID uint64,
	appRolloutRelationType entity.AppRolloutRelationType,
	input CreateRolloutInput,
) (entity.Rollout, *errs.Error) {
	rollout := entity.Rollout{
		SelectorID:  input.VersionSelectorID,
		ActivatorID: input.ActivatorID,
		IsEnabled:   input.IsEnabled,
		Viewers:     0,
	}
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		err := r.rolloutDao.CreateRollout(ct, tx, rollout)
		if err != nil {
			return err
		}

		return r.appRolloutRelationDao.CreateAppRolloutRelation(ct, tx, entity.AppRolloutRelation{
			AppID:     appID,
			RolloutID: rollout.ID,
			Type:      appRolloutRelationType,
		})
	})
	return rollout, err
}

func (r *Rollout) DeleteRollout(ct context.Context, rolloutID uint64) (entity.Rollout, *errs.Error) {
	var rollout entity.Rollout
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		rollout, err = r.rolloutDao.FindRolloutByIDWithTx(ct, tx, rolloutID)
		if err != nil {
			return err
		}

		return r.rolloutDao.DeleteRollout(ct, tx, rolloutID)
	})

	return rollout, err
}

func (r *Rollout) UpdateRollout(ct context.Context, rolloutID uint64, input UpdateRolloutInput) (entity.Rollout, *errs.Error) {
	var rollout entity.Rollout
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		rollout, err = r.rolloutDao.FindRolloutByIDWithTx(ct, tx, rolloutID)
		if err != nil {
			return err
		}

		rollout.SelectorID = input.VersionSelectorID
		rollout.ActivatorID = input.ActivatorID
		rollout.IsEnabled = input.IsEnabled
		return r.rolloutDao.UpdateRollout(ct, tx, rollout)
	})

	return rollout, err
}

func (r *Rollout) FindRolloutsByGroupID(ct context.Context, groupID uint64) ([]entity.Rollout, *errs.Error) {
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)

	var rollouts []entity.Rollout
	err := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		groupRolloutRelations, err := r.groupRolloutRelationDao.FindGroupRolloutRelationsByGroupID(ct, groupID)
		if err != nil {
			return err
		}

		rolloutIDs := collect.Map(groupRolloutRelations, func(groupRolloutRelation entity.GroupRolloutRelation, index int) uint64 {
			return groupRolloutRelation.RolloutID
		})
		rollouts, err = r.rolloutDao.FindRolloutsByIDs(ct, rolloutIDs)
		return err
	})

	return rollouts, err
}

func (r *Rollout) GetActiveAppVersionNumberForTeam(ct context.Context, appTeamInstallation entity.TeamAppInstallation) (*int, *errs.Error) {
	teamID := appTeamInstallation.InstalledTeamID
	appID := appTeamInstallation.AppID
	teamGroupsIDs, err := r.appGroupRelationDao.FindGroupIDsByAppIDAndRelationType(ct, appID, entity.AppGroupRelationTypeTeam)
	if err != nil {
		return nil, err
	}

	var maxActiveVersionNumber int = math.MinInt
	for _, teamGroupID := range teamGroupsIDs {
		versionNumber, err := r.getTeamGroupActiveVersion(ct, teamID, teamGroupID)
		if err != nil {
			return nil, err
		}

		if *versionNumber > maxActiveVersionNumber {
			maxActiveVersionNumber = *versionNumber
		}
	}

	return &maxActiveVersionNumber, nil
}

func (r *Rollout) newRollout(ct context.Context, rawRollout entity.Rollout) (rollout.Rollout, *errs.Error) {
	activatorID := rawRollout.ActivatorID
	versionSelector, err := r.getRolloutVersionSelector(ct, rawRollout.ID, rawRollout.SelectorID)
	if err != nil {
		return rollout.Rollout{}, err
	}

	activator, err := r.getRolloutActivator(ct, rawRollout.ID, activatorID)
	if err != nil {
		return rollout.Rollout{}, err
	}

	rolloutStore := store.NewRollout(r.logger, r.transactionFactory, r.stateSyncer, r.rolloutDao, rawRollout.ID)
	return rollout.NewRollout(ct, rolloutStore, activator, versionSelector)
}

func (r *Rollout) FindActivatorByID(ct context.Context, activatorID uint64) (entity.ActivatorUnion, *errs.Error) {
	return r.activatorRepository.FindActivatorByID(ct, activatorID)
}

func (r *Rollout) CreateTimeRangeActivator(ct context.Context, startAt *time.Time, endAt *time.Time) (entity.TimeRangeActivator, *errs.Error) {
	timeRangeActivator := entity.TimeRangeActivator{
		StartAt: startAt,
		EndAt:   endAt,
		Activator: entity.Activator{
			CreatedAt: time.Now().UTC(),
		},
	}
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		return r.activatorRepository.CreateTimeRangeActivator(ct, tx, timeRangeActivator)
	})
	return timeRangeActivator, err
}

func (r *Rollout) CreateMaxViewersActivator(ct context.Context, maxViewers int) (entity.MaxViewersActivator, *errs.Error) {
	maxViewersActivator := entity.MaxViewersActivator{
		MaxViewers: maxViewers,
		Activator: entity.Activator{
			CreatedAt: time.Now().UTC(),
		},
	}
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		return r.activatorRepository.CreateMaxViewersActivator(ct, tx, maxViewersActivator)
	})
	return maxViewersActivator, err
}

func (r *Rollout) CreatePercentageActivator(ct context.Context, percentage int) (entity.PercentageActivator, *errs.Error) {
	percentageActivator := entity.PercentageActivator{
		Percentage: percentage,
		Activator: entity.Activator{
			CreatedAt: time.Now().UTC(),
		},
	}
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		return r.activatorRepository.CreatePercentageActivator(ct, tx, percentageActivator)
	})
	return percentageActivator, err
}

func (r *Rollout) FindVersionSelectorByID(ct context.Context, selectorID uint64) (entity.VersionSelectorUnion, *errs.Error) {
	versionSelectorUnion := entity.VersionSelectorUnion{}
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)

	txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		versionSelectorUnion, err = r.versionSelectorRepository.FindVersionSelectorByID(ct, tx, selectorID)
		if err != nil {
			return err
		}

		return nil
	})

	return versionSelectorUnion, nil
}

func (r *Rollout) CreateStaticVersionSelector(ct context.Context, versionNumber int) (entity.StaticVersionSelector, *errs.Error) {
	staticVersionSelector := entity.StaticVersionSelector{}
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		staticVersionSelector, err = r.versionSelectorRepository.CreateStaticVersionSelector(ct, tx, versionNumber)
		if err != nil {
			return err
		}

		return nil
	})
	return staticVersionSelector, err
}

func (r *Rollout) CreateExperimentVersionSelector(ct context.Context, versionNumbers []int) (entity.ExperimentVersionSelector, *errs.Error) {
	experimentVersionSelector := entity.ExperimentVersionSelector{}
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		experimentVersionSelector, err = r.versionSelectorRepository.CreateExperimentVersionSelector(ct, tx, versionNumbers)
		if err != nil {
			return err
		}

		return nil
	})
	return experimentVersionSelector, err
}

func (r *Rollout) getTeamGroupActiveVersion(ct context.Context, teamID uint64, groupID uint64) (*int, *errs.Error) {
	groupRolloutRelations, err := r.groupRolloutRelationDao.FindGroupRolloutRelationsByGroupID(ct, groupID)
	if err != nil {
		return nil, err
	}

	sort.Slice(groupRolloutRelations, func(i, j int) bool {
		return groupRolloutRelations[i].OrderIndex < groupRolloutRelations[j].OrderIndex
	})
	sortedRolloutIDs := collect.Map(groupRolloutRelations, func(groupRolloutRelation entity.GroupRolloutRelation, index int) uint64 {
		return groupRolloutRelation.RolloutID
	})
	rawRollouts, err := r.rolloutDao.FindRolloutsByIDs(ct, sortedRolloutIDs)
	if err != nil {
		return nil, err
	}

	rollouts := make([]rollout.Rollout, 0)
	for _, rawRollout := range rawRollouts {
		rollout, err := r.newRollout(ct, rawRollout)
		if err != nil {
			return nil, err
		}

		rollouts = append(rollouts, rollout)
	}

	orderedRollouts := rollout.OrderedRollouts(rollouts)
	versionNumber, err := orderedRollouts.GetVersionNumber(ct, teamID)
	return versionNumber, err
}

// func (r *Rollout) createVersionSelector(ct context.Context, versionSelectorType entity.VersionSelectorType, versionNumbers []int) (entity.VersionSelectorUnion, *errs.Error) {
// 	versionSelectorUnion := entity.VersionSelectorUnion{}
// 	txCtx := transaction.NewTransactionsContext(
// 		r.logger,
// 		r.transactionFactory,
// 		r.stateSyncer,
// 		ct,
// 	)
// 	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
// 		return r.versionSelectorRepository.CreateVersionSelector(ct, tx, versionSelectorType, versionNumbers)
// 	})
// 	return versionSelectorUnion, err
// }

func (r *Rollout) getRolloutVersionSelector(ct context.Context, rolloutID uint64, selectorID uint64) (rollout.VersionSelector, *errs.Error) {
	var versionSelector rollout.VersionSelector
	rawVersionSelector, err := r.FindVersionSelectorByID(ct, selectorID)
	if err != nil {
		return nil, err
	}

	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)

	err = txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		versionSelectorUnion, err := r.versionSelectorRepository.FindVersionSelectorByID(ct, tx, selectorID)
		if err != nil {
			return err
		}

		switch rawVersionSelector.Type {
		case entity.VersionSelectorTypeStatic:
			versionSelector = rollout.NewStaticVersionSelector(versionSelectorUnion.StaticVersionSelector.VersionNumber)
		case entity.VersionSelectorTypeExperiment:
			store := store.NewExperimentVersionSelector(r.logger, r.transactionFactory, r.stateSyncer, r.rolloutViewerDao, rolloutID)
			versionSelector = rollout.NewExperimentVersionSelector(store, randgen.NewBuiltinRanGen(), versionSelectorUnion.ExperimentVersionSelector.VersionNumbers)
		default:
			return errs.NewError(errs.Unknown, fmt.Sprintf("unknown version selector type %s", rawVersionSelector.Type))
		}

		return nil
	})

	return versionSelector, err
}

func (r *Rollout) getRolloutActivator(ct context.Context, rolloutID uint64, activatorID uint64) (rollout.Activator, *errs.Error) {
	var activator rollout.Activator
	activatorUnion, err := r.activatorRepository.FindActivatorByID(ct, activatorID)
	if err != nil {
		return nil, err
	}

	switch activatorUnion.Type {
	case entity.ActivatorTypeTimeRange:
		rawActivator := activatorUnion.TimeRangeActivator
		activator = rollout.NewTimeRangeActivator(clock.New(), rawActivator.StartAt, rawActivator.EndAt)

		if err != nil {
			return nil, err
		}
	case entity.ActivatorTypeMaxViewers:
		rawActivator := activatorUnion.MaxViewersActivator

		activatorStore := store.NewMaxViewersActivator(r.logger, r.transactionFactory, r.stateSyncer, r.rolloutViewerDao, r.rolloutDao, rolloutID)
		activator, err = rollout.NewMaxViewersActivator(ct, activatorStore, rawActivator.MaxViewers)
		if err != nil {
			return nil, err
		}
	case entity.ActivatorTypePercentage:
		rawActivator := activatorUnion.PercentageActivator

		store := store.NewPercentageActivator(r.logger, r.transactionFactory, r.stateSyncer, r.rolloutViewerDao, rolloutID)
		activator = rollout.NewPercentageActivator(store, randgen.NewBuiltinRanGen(), rawActivator.Percentage)
	default:
		return nil, errs.NewError(errs.Unknown, fmt.Sprintf("unknown activator type %s", activatorUnion.Type))
	}

	return activator, nil
}

func NewRollout(
	logger telemetry.Logger,
	transactionFactory cloudTransaction.Factory,
	stateSyncer *realtime.StateSyncer,
	appGroupRelationDao dao.AppGroupRelation,
	rolloutDao dao.Rollout,
	rolloutViewerDao dao.RolloutViewer,
	groupRolloutRelationDao dao.GroupRolloutRelation,
	appRolloutRelationDao dao.AppRolloutRelation,
	appVersionDao dao.AppVersion,
	versionSelectorRepository *repository.VersionSelector,
	activatorRepository *repository.Activator,
) *Rollout {
	return &Rollout{
		logger:                    logger,
		transactionFactory:        transactionFactory,
		stateSyncer:               stateSyncer,
		appGroupRelationDao:       appGroupRelationDao,
		rolloutDao:                rolloutDao,
		rolloutViewerDao:          rolloutViewerDao,
		groupRolloutRelationDao:   groupRolloutRelationDao,
		appRolloutRelationDao:     appRolloutRelationDao,
		appVersionDao:             appVersionDao,
		versionSelectorRepository: versionSelectorRepository,
		activatorRepository:       activatorRepository,
	}
}
