package service

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/randgen"
	"github.com/teamyapp/cloud/libs/rollout"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type Rollout struct {
	logger                    telemetry.Logger
	cloudClientRegistry       *client.Registry
	transactionFactory        transaction.Factory
	stateSyncer               *realtime.StateSyncer
	appGroupRelationDao       dao.AppGroupRelation
	rolloutDao                dao.Rollout
	rolloutStoreDao           dao.RolloutStore
	rolloutViewerDao          dao.RolloutViewer
	groupRolloutRelationDao   dao.GroupRolloutRelation
	appRolloutRelationDao     dao.AppRolloutRelation
	rolloutVersionRelationDao dao.RolloutVersionRelation
	timeRangeActivatorDao     dao.TimeRangeActivator
	maxViewersActivatorDao    dao.MaxViewersActivator
	percentageActivatorDao    dao.PercentageActivator
	appVersionDao             dao.AppVersion
}

func (r *Rollout) FindUserRolloutsByAppID(ct context.Context, appID uint64) ([]entity.Rollout, *errs.Error) {
	rolloutIDs, err := r.appRolloutRelationDao.FindRolloutIDsByAppIDAndType(ct, appID, entity.AppRolloutRelationTypeUser)
	if err != nil {
		return nil, err
	}

	return r.rolloutDao.FindRolloutsByIDs(ct, rolloutIDs)
}

func (r *Rollout) FindTeamRolloutsByAppID(ct context.Context, appID uint64) ([]entity.Rollout, *errs.Error) {
	rolloutIDs, err := r.appRolloutRelationDao.FindRolloutIDsByAppIDAndType(ct, appID, entity.AppRolloutRelationTypeTeam)
	if err != nil {
		return nil, err
	}

	return r.rolloutDao.FindRolloutsByIDs(ct, rolloutIDs)
}

func (r *Rollout) CreateRollout(
	ct context.Context,
	selectorType entity.SelectorType,
	activatorID uint64,
	activatorType entity.ActivatorType,
	isEnabled bool,
) (entity.Rollout, *errs.Error) {
	genSelectorIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "selectorID"}
	genSelectorIDRes, rpcErr := r.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genSelectorIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.Rollout{}, internalErr
	}

	rollout := entity.Rollout{
		SelectorID:    genSelectorIDRes.UniqueNumber,
		SelectorType:  selectorType,
		ActivatorID:   activatorID,
		ActivatorType: activatorType,
		IsEnabled:     isEnabled,
	}

	return r.rolloutDao.CreateRollout(ct, rollout)
}

func (r *Rollout) DeleteRollout(ct context.Context, rolloutID uint64) (entity.Rollout, *errs.Error) {
	var rollout entity.Rollout
	txCtx := TransactionsContext{
		logger:             r.logger,
		transactionFactory: r.transactionFactory,
		stateSyncer:        r.stateSyncer,
		ct:                 ct,
	}

	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		rollout, err = r.rolloutDao.FindRolloutByIDWithTx(ct, tx, rolloutID)
		if err != nil {
			return err
		}

		return r.rolloutDao.DeleteRolloutWithTx(ct, tx, rolloutID)
	})

	if err != nil {
		return entity.Rollout{}, err
	}

	return rollout, nil

}

func (r *Rollout) FindRolloutsByGroupID(ct context.Context, groupID uint64) ([]entity.Rollout, *errs.Error) {
	rolloutIDs, err := r.groupRolloutRelationDao.FindRolloutIDsByGroupIDAndSortByOrderedIndex(ct, groupID)
	if err != nil {
		return nil, err
	}

	return r.rolloutDao.FindRolloutsByIDs(ct, rolloutIDs)
}

func (r *Rollout) GetActiveAppVersionNumberForTeam(ct context.Context, appTeamInstallation entity.TeamAppInstallation) (int, *errs.Error) {
	teamID := appTeamInstallation.InstalledTeamID
	appID := appTeamInstallation.AppID
	teamGroupsIDs, err := r.appGroupRelationDao.FindGroupIDsByAppIDAndType(ct, appID, entity.AppGroupRelationTypeTeam)
	sortedRolloutIDs := make([]uint64, 0)
	for _, teamGroupID := range teamGroupsIDs {
		rolloutIDs, err := r.groupRolloutRelationDao.FindRolloutIDsByGroupIDAndSortByOrderedIndex(ct, teamGroupID)
		if err != nil {
			return 0, err
		}

		sortedRolloutIDs = append(sortedRolloutIDs, rolloutIDs...)
	}

	rolloutEntities, err := r.rolloutDao.FindRolloutsByIDs(ct, sortedRolloutIDs)
	if err != nil {
		return 0, err
	}

	rollouts := make([]rollout.Rollout, 0)
	for _, rolloutEntity := range rolloutEntities {
		rollout, err := r.newRollout(ct, rolloutEntity)
		if err != nil {
			return 0, err
		}

		rollouts = append(rollouts, rollout)
	}

	orderedRollouts := rollout.OrderedRollouts(rollouts)
	versionNumber, err := orderedRollouts.GetVersionNumber(ct, teamID)
	if err != nil {
		return 0, err
	}

	return *versionNumber, nil
}

func (r *Rollout) newRollout(ct context.Context, rolloutEntity entity.Rollout) (rollout.Rollout, *errs.Error) {
	activatorType := rolloutEntity.ActivatorType
	activatorID := rolloutEntity.ActivatorID
	selectorType := rolloutEntity.SelectorType
	versionNumbers, err := r.rolloutVersionRelationDao.FindVersionNumbersByRolloutID(ct, rolloutEntity.ID)
	if err != nil {
		return rollout.Rollout{}, err
	}

	versionSelector, err := r.getRolloutVersionSelector(ct, rolloutEntity.ID, versionNumbers, selectorType)
	if err != nil {
		return rollout.Rollout{}, err
	}

	activator, err := r.getRolloutVersionActivator(ct, rolloutEntity.ID, activatorID, activatorType)
	if err != nil {
		return rollout.Rollout{}, err
	}

	rolloutStore := NewRolloutStore(r.rolloutDao, rolloutEntity.ID)
	return rollout.NewRollout(ct, rolloutStore, activator, versionSelector)
}

func (r *Rollout) CreateTimeRangeActivator(ct context.Context, timeRangeActivator entity.TimeRangeActivator) (entity.TimeRangeActivator, *errs.Error) {
	return r.timeRangeActivatorDao.CreateTimeRangeActivator(ct, timeRangeActivator)
}

func (r *Rollout) FindTimeRangeActivatorByID(ct context.Context, activatorID uint64) (entity.TimeRangeActivator, *errs.Error) {
	return r.timeRangeActivatorDao.FindTimeRangeActivatorByID(ct, activatorID)
}

func (r *Rollout) CreateMaxViewersActivator(ct context.Context, maxViewersActivator entity.MaxViewersActivator) (entity.MaxViewersActivator, *errs.Error) {
	return r.maxViewersActivatorDao.CreateMaxViewersActivator(ct, maxViewersActivator)
}

func (r *Rollout) FindMaxViewersActivatorByID(ct context.Context, activatorID uint64) (entity.MaxViewersActivator, *errs.Error) {
	return r.maxViewersActivatorDao.FindMaxViewersActivatorByID(ct, activatorID)
}

func (r *Rollout) CreatePercentageActivator(ct context.Context, activator entity.PercentageActivator) (entity.PercentageActivator, *errs.Error) {
	return r.percentageActivatorDao.CreatePercentageActivator(ct, activator)
}

func (r *Rollout) FindPercentageActivatorByID(ct context.Context, activatorID uint64) (entity.PercentageActivator, *errs.Error) {
	return r.percentageActivatorDao.FindPercentageActivatorByID(ct, activatorID)
}

func (r *Rollout) getRolloutVersionSelector(ct context.Context, rolloutID uint64, versionNumbers []int, selectorType entity.SelectorType) (rollout.VersionSelector, *errs.Error) {
	var versionSelector rollout.VersionSelector
	if len(versionNumbers) == 0 {
		return nil, errs.NewError(errs.NotFound, "app version not found")
	}

	switch selectorType {
	case entity.SelectorTypeStatic:
		if len(versionNumbers) > 1 {
			r.logger.WarningWithContext(ct, "multiple app versions found for a static selector, use the first one")
		}

		versionSelector = rollout.NewStaticVersionSelector(versionNumbers[0])
	case entity.SelectorTypeExperiment:
		store := NewExperimentVersionSelectorStore(r.rolloutViewerDao, rolloutID)
		versionSelector = rollout.NewExperimentVersionSelector(store, randgen.NewBuiltinRanGen(), versionNumbers)
	}

	return versionSelector, nil
}

func (r *Rollout) getRolloutVersionActivator(ct context.Context, rolloutID uint64, activatorID uint64, activatorType entity.ActivatorType) (rollout.Activator, *errs.Error) {
	var activator rollout.Activator
	switch activatorType {
	case entity.ActivatorTypeTimeRange:
		activatorEntity, err := r.FindTimeRangeActivatorByID(ct, activatorID)
		activator = rollout.NewTimeRangeActivator(clock.New(), &activatorEntity.StartAt, &activatorEntity.EndAt)

		if err != nil {
			return nil, err
		}
	case entity.ActivatorTypeMaxViewers:
		activatorEntity, err := r.FindMaxViewersActivatorByID(ct, activatorID)
		if err != nil {
			return nil, err
		}

		store := NewMaxViewersActivatorStore(r.rolloutViewerDao, r.rolloutStoreDao, rolloutID)
		activator, err = rollout.NewMaxViewersActivator(ct, store, activatorEntity.MaxViewers)
		if err != nil {
			return nil, err
		}
	case entity.ActivatorTypePercentage:
		activatorEntity, err := r.FindPercentageActivatorByID(ct, activatorID)
		if err != nil {
			return nil, err
		}

		store := NewPercentageActivatorStore(r.rolloutViewerDao, rolloutID)
		activator = rollout.NewPercentageActivator(store, randgen.NewBuiltinRanGen(), activatorEntity.Percentage)
	}

	return activator, nil
}

func NewRollout(
	logger telemetry.Logger,
	cloudClientRegistry *client.Registry,
	transactionFactory transaction.Factory,
	stateSyncer *realtime.StateSyncer,
	appGroupRelationDao dao.AppGroupRelation,
	rolloutDao dao.Rollout,
	rolloutStoreDao dao.RolloutStore,
	rolloutViewerDao dao.RolloutViewer,
	groupRolloutRelationDao dao.GroupRolloutRelation,
	appRolloutRelationDao dao.AppRolloutRelation,
	rolloutVersionRelationDao dao.RolloutVersionRelation,
	timeRangeActivatorDao dao.TimeRangeActivator,
	maxViewersActivatorDao dao.MaxViewersActivator,
	percentageActivatorDao dao.PercentageActivator,
	appVersionDao dao.AppVersion,
) *Rollout {
	return &Rollout{
		logger:                    logger,
		cloudClientRegistry:       cloudClientRegistry,
		transactionFactory:        transactionFactory,
		stateSyncer:               stateSyncer,
		appGroupRelationDao:       appGroupRelationDao,
		rolloutDao:                rolloutDao,
		rolloutStoreDao:           rolloutStoreDao,
		rolloutViewerDao:          rolloutViewerDao,
		groupRolloutRelationDao:   groupRolloutRelationDao,
		appRolloutRelationDao:     appRolloutRelationDao,
		rolloutVersionRelationDao: rolloutVersionRelationDao,
		timeRangeActivatorDao:     timeRangeActivatorDao,
		maxViewersActivatorDao:    maxViewersActivatorDao,
		percentageActivatorDao:    percentageActivatorDao,
		appVersionDao:             appVersionDao,
	}
}
