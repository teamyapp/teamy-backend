package service

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/randgen"
	"github.com/teamyapp/cloud/libs/rollout"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
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
	logger                            telemetry.Logger
	cloudClientRegistry               *client.Registry
	transactionFactory                transaction.Factory
	stateSyncer                       *realtime.StateSyncer
	appGroupRelationDao               dao.AppGroupRelation
	rolloutDao                        dao.Rollout
	rolloutStoreDao                   dao.RolloutStore
	rolloutViewerDao                  dao.RolloutViewer
	groupRolloutRelationDao           dao.GroupRolloutRelation
	appRolloutRelationDao             dao.AppRolloutRelation
	versionSelectorVersionRelationDao dao.VersionSelectorVersionRelation
	timeRangeActivatorDao             dao.TimeRangeActivator
	maxViewersActivatorDao            dao.MaxViewersActivator
	percentageActivatorDao            dao.PercentageActivator
	activatorTypeRelationDao          dao.ActivatorTypeRelation
	versionSelectorDao                dao.VersionSelector
	appVersionDao                     dao.AppVersion
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
	}
	rollout, err := r.rolloutDao.CreateRollout(ct, rollout)
	if err != nil {
		return entity.Rollout{}, err
	}

	_, err = r.appRolloutRelationDao.CreateAppRolloutRelation(ct, entity.AppRolloutRelation{
		AppID:     appID,
		RolloutID: rollout.ID,
		Type:      appRolloutRelationType,
	})
	return rollout, err
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

	return rollout, err
}

func (r *Rollout) UpdateRollout(ct context.Context, rolloutID uint64, input UpdateRolloutInput) (entity.Rollout, *errs.Error) {
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

		rollout.SelectorID = input.VersionSelectorID
		rollout.ActivatorID = input.ActivatorID
		rollout.IsEnabled = input.IsEnabled
		return r.rolloutDao.UpdateRolloutWithTx(ct, tx, rollout)
	})

	return rollout, err
}

func (r *Rollout) FindRolloutsByGroupID(ct context.Context, groupID uint64) ([]entity.Rollout, *errs.Error) {
	rolloutIDs, err := r.groupRolloutRelationDao.FindRolloutIDsByGroupID(ct, groupID)
	if err != nil {
		return nil, err
	}

	return r.rolloutDao.FindRolloutsByIDs(ct, rolloutIDs)
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

func (r *Rollout) FindVersionNumbersByVersionSelectorID(ct context.Context, rolloutID uint64) ([]int, *errs.Error) {
	return r.versionSelectorVersionRelationDao.FindVersionNumbersBySelectorID(ct, rolloutID)
}

func (r *Rollout) newRollout(ct context.Context, rolloutEntity entity.Rollout) (rollout.Rollout, *errs.Error) {
	activatorID := rolloutEntity.ActivatorID
	versionNumbers, err := r.FindVersionNumbersByVersionSelectorID(ct, rolloutEntity.SelectorID)
	if err != nil {
		return rollout.Rollout{}, err
	}

	versionSelector, err := r.getRolloutVersionSelector(ct, rolloutEntity.ID, rolloutEntity.SelectorID, versionNumbers)
	if err != nil {
		return rollout.Rollout{}, err
	}

	activator, err := r.getRolloutVersionActivator(ct, rolloutEntity.ID, activatorID)
	if err != nil {
		return rollout.Rollout{}, err
	}

	rolloutStore := NewRolloutStore(r.rolloutDao, rolloutEntity.ID)
	return rollout.NewRollout(ct, rolloutStore, activator, versionSelector)
}

func (r *Rollout) CreateTimeRangeActivator(ct context.Context, startAt *time.Time, endAt *time.Time) (entity.TimeRangeActivator, *errs.Error) {
	timeRangeActivator := entity.TimeRangeActivator{
		StartAt: startAt,
		EndAt:   endAt,
		Activator: entity.Activator{
			CreatedAt: time.Now().UTC(),
		},
	}

	return r.timeRangeActivatorDao.CreateTimeRangeActivator(ct, timeRangeActivator)
}

func (r *Rollout) FindTimeRangeActivatorByID(ct context.Context, activatorID uint64) (entity.TimeRangeActivator, *errs.Error) {
	return r.timeRangeActivatorDao.FindTimeRangeActivatorByID(ct, activatorID)
}

func (r *Rollout) CreateMaxViewersActivator(ct context.Context, maxViewers int) (entity.MaxViewersActivator, *errs.Error) {
	maxViewersActivator := entity.MaxViewersActivator{
		MaxViewers: maxViewers,
		Activator: entity.Activator{
			CreatedAt: time.Now().UTC(),
		},
	}
	return r.maxViewersActivatorDao.CreateMaxViewersActivator(ct, maxViewersActivator)
}

func (r *Rollout) FindMaxViewersActivatorByID(ct context.Context, activatorID uint64) (entity.MaxViewersActivator, *errs.Error) {
	return r.maxViewersActivatorDao.FindMaxViewersActivatorByID(ct, activatorID)
}

func (r *Rollout) CreatePercentageActivator(ct context.Context, percentage int) (entity.PercentageActivator, *errs.Error) {
	percentageActivator := entity.PercentageActivator{
		Percentage: percentage,
		Activator: entity.Activator{
			CreatedAt: time.Now().UTC(),
		},
	}
	return r.percentageActivatorDao.CreatePercentageActivator(ct, percentageActivator)
}

func (r *Rollout) FindPercentageActivatorByID(ct context.Context, activatorID uint64) (entity.PercentageActivator, *errs.Error) {
	return r.percentageActivatorDao.FindPercentageActivatorByID(ct, activatorID)
}

func (r *Rollout) FindVersionSelectorByID(ct context.Context, selectorID uint64) (entity.VersionSelector, *errs.Error) {
	return r.versionSelectorDao.FindVersionSelectorByID(ct, selectorID)
}

func (r *Rollout) CreateStaticVersionSelector(ct context.Context, versionNumber int) (entity.VersionSelector, *errs.Error) {
	return r.createVersionSelector(ct, entity.VersionSelectorTypeStatic, []int{versionNumber})
}

func (r *Rollout) CreateExperimentVersionSelector(ct context.Context, versionNumbers []int) (entity.VersionSelector, *errs.Error) {
	return r.createVersionSelector(ct, entity.VersionSelectorTypeExperiment, versionNumbers)
}

func (r *Rollout) FindActivatorTypeByID(ct context.Context, activatorID uint64) (entity.ActivatorType, *errs.Error) {
	return r.activatorTypeRelationDao.FindActivatorTypeByID(ct, activatorID)
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
	rolloutEntities, err := r.rolloutDao.FindRolloutsByIDs(ct, sortedRolloutIDs)
	if err != nil {
		return nil, err
	}

	rollouts := make([]rollout.Rollout, 0)
	for _, rolloutEntity := range rolloutEntities {
		rollout, err := r.newRollout(ct, rolloutEntity)
		if err != nil {
			return nil, err
		}

		rollouts = append(rollouts, rollout)
	}

	orderedRollouts := rollout.OrderedRollouts(rollouts)
	versionNumber, err := orderedRollouts.GetVersionNumber(ct, teamID)
	return versionNumber, err
}

func (r *Rollout) createVersionSelector(ct context.Context, versionSelectorType entity.VersionSelectorType, versionNumbers []int) (entity.VersionSelector, *errs.Error) {
	genSelectorIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "selectorID"}
	genSelectorIDRes, rpcErr := r.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genSelectorIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.VersionSelector{}, internalErr
	}

	versionSelector := entity.VersionSelector{
		ID:   genSelectorIDRes.UniqueNumber,
		Type: versionSelectorType,
	}

	versionSelector, err := r.versionSelectorDao.CreateVersionSelector(ct, versionSelector)
	if err != nil {
		return entity.VersionSelector{}, err
	}

	for _, versionNumber := range versionNumbers {
		_, err := r.versionSelectorVersionRelationDao.CreateVersionSelectorVersionRelation(ct, entity.VersionSelectorVersionRelation{
			VersionSelectorID: versionSelector.ID,
			VersionNumber:     versionNumber,
		})
		if err != nil {
			return entity.VersionSelector{}, err
		}
	}

	return versionSelector, nil
}

func (r *Rollout) getRolloutVersionSelector(ct context.Context, rolloutID uint64, selectorID uint64, versionNumbers []int) (rollout.VersionSelector, *errs.Error) {
	var versionSelector rollout.VersionSelector
	if len(versionNumbers) == 0 {
		return nil, errs.NewError(errs.NotFound, "app version not found")
	}

	versionSelectorEntity, err := r.FindVersionSelectorByID(ct, selectorID)
	if err != nil {
		return nil, err
	}

	switch versionSelectorEntity.Type {
	case entity.VersionSelectorTypeStatic:
		if len(versionNumbers) > 1 {
			r.logger.WarningWithContext(ct, "multiple app versions found for a static selector, use the first one")
		}

		versionSelector = rollout.NewStaticVersionSelector(versionNumbers[0])
	case entity.VersionSelectorTypeExperiment:
		store := NewExperimentVersionSelectorStore(r.rolloutViewerDao, rolloutID)
		versionSelector = rollout.NewExperimentVersionSelector(store, randgen.NewBuiltinRanGen(), versionNumbers)
	}

	return versionSelector, nil
}

func (r *Rollout) getRolloutVersionActivator(ct context.Context, rolloutID uint64, activatorID uint64) (rollout.Activator, *errs.Error) {
	var activator rollout.Activator
	activatorType, err := r.activatorTypeRelationDao.FindActivatorTypeByID(ct, activatorID)
	if err != nil {
		return nil, err
	}

	switch activatorType {
	case entity.ActivatorTypeTimeRange:
		activatorEntity, err := r.FindTimeRangeActivatorByID(ct, activatorID)
		activator = rollout.NewTimeRangeActivator(clock.New(), activatorEntity.StartAt, activatorEntity.EndAt)

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
	versionSelectorVersionRelationDao dao.VersionSelectorVersionRelation,
	timeRangeActivatorDao dao.TimeRangeActivator,
	maxViewersActivatorDao dao.MaxViewersActivator,
	percentageActivatorDao dao.PercentageActivator,
	activatorTypeRelationDao dao.ActivatorTypeRelation,
	versionSelectorDao dao.VersionSelector,
	appVersionDao dao.AppVersion,
) *Rollout {
	return &Rollout{
		logger:                            logger,
		cloudClientRegistry:               cloudClientRegistry,
		transactionFactory:                transactionFactory,
		stateSyncer:                       stateSyncer,
		appGroupRelationDao:               appGroupRelationDao,
		rolloutDao:                        rolloutDao,
		rolloutStoreDao:                   rolloutStoreDao,
		rolloutViewerDao:                  rolloutViewerDao,
		groupRolloutRelationDao:           groupRolloutRelationDao,
		appRolloutRelationDao:             appRolloutRelationDao,
		versionSelectorVersionRelationDao: versionSelectorVersionRelationDao,
		timeRangeActivatorDao:             timeRangeActivatorDao,
		maxViewersActivatorDao:            maxViewersActivatorDao,
		percentageActivatorDao:            percentageActivatorDao,
		activatorTypeRelationDao:          activatorTypeRelationDao,
		versionSelectorDao:                versionSelectorDao,
		appVersionDao:                     appVersionDao,
	}
}
