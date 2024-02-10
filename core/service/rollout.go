package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/delta"
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
	Name              string
	VersionSelectorID uint64
	ActivatorID       uint64
	IsEnabled         bool
	GroupIDs          []uint64
}

type UpdateRolloutInput struct {
	Name              string
	VersionSelectorID uint64
	ActivatorID       uint64
	IsEnabled         bool
	GroupIDs          []uint64
}

type UpdateActivatorInput struct {
	Type       entity.ActivatorType
	StartAt    *time.Time
	EndAt      *time.Time
	MaxViewers *int32
	Percentage *int32
}

type UpdateVersionSelectorInput struct {
	Type           entity.VersionSelectorType
	VersionNumber  *int32
	VersionNumbers []int
}

type Rollout struct {
	logger                    telemetry.Logger
	cloudClientRegistry       *client.Registry
	transactionFactory        cloudTransaction.Factory
	stateSyncer               *realtime.StateSyncer
	appGroupRelationDao       dao.AppGroupRelation
	rolloutDao                dao.Rollout
	rolloutViewerDao          dao.RolloutViewer
	groupRolloutRelationDao   dao.GroupRolloutRelation
	appRolloutRelationDao     dao.AppRolloutRelation
	groupMemberRelationDao    dao.GroupMemberRelation
	appVersionDao             dao.AppVersion
	versionSelectorRepository *repository.VersionSelector
	activatorRepository       *repository.Activator
	groupRepository           *repository.Group
	activatorDao              dao.Activator
	teamDao                   dao.Team
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
	genRolloutIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "rolloutID"}
	genRolloutRes, rpcErr := r.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genRolloutIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.Rollout{}, internalErr
	}

	rollout := entity.Rollout{
		ID:          genRolloutRes.UniqueNumber,
		Name:        input.Name,
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

		for _, groupID := range input.GroupIDs {
			maxRolloutIndex, err := r.groupRepository.UpdateMaxRolloutIndexByGroupID(ct, tx, groupID, 1)
			if err != nil {
				return err
			}

			groupRolloutRelation := entity.GroupRolloutRelation{
				GroupID:    groupID,
				RolloutID:  rollout.ID,
				OrderIndex: maxRolloutIndex,
			}
			err = r.groupRolloutRelationDao.CreateGroupRolloutRelation(ct, tx, groupRolloutRelation)
			if err != nil {
				return err
			}
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

		_, err = r.activatorRepository.DeleteActivator(ct, tx, rollout.ActivatorID)
		if err != nil {
			return nil
		}

		err = r.versionSelectorRepository.DeleteVersionSelector(ct, tx, rollout.SelectorID)
		if err != nil {
			return nil
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

		newGroupIDSet := map[uint64]bool{}
		for _, groupID := range input.GroupIDs {
			newGroupIDSet[groupID] = true
		}

		currentGroupIDSet := map[uint64]bool{}
		oldGroupRelations, err := r.groupRolloutRelationDao.FindGroupRolloutRelationsByRolloutIDWithTx(ct, tx, rolloutID)
		if err != nil {
			return err
		}

		for _, groupRelation := range oldGroupRelations {
			currentGroupIDSet[groupRelation.GroupID] = true
		}

		detected := delta.DetectMapDelta(
			currentGroupIDSet,
			newGroupIDSet,
			delta.DetectValueDelta[bool],
			delta.ToValueDelta[bool],
		)

		for groupID, detectedValue := range detected.Value {
			switch detectedValue.KeyStatus {
			case delta.AddedStatus:
				maxRolloutIndex, err := r.groupRepository.UpdateMaxRolloutIndexByGroupID(ct, tx, groupID, 1)
				if err != nil {
					return err
				}

				err = r.groupRolloutRelationDao.CreateGroupRolloutRelation(ct, tx, entity.GroupRolloutRelation{
					GroupID:    groupID,
					RolloutID:  rolloutID,
					OrderIndex: maxRolloutIndex,
				})
				if err != nil {
					return err
				}

			case delta.RemovedStatus:
				relation, err := r.groupRolloutRelationDao.FindGroupRolloutByGroupIDAndRolloutIDWithTx(ct, tx, groupID, rolloutID)
				if err != nil {
					return err
				}

				err = r.groupRolloutRelationDao.DeleteGroupRolloutRelationsByGroupIDAndRolloutID(ct, tx, groupID, rolloutID)
				if err != nil {
					return err
				}

				err = r.groupRolloutRelationDao.UpdateFromOrderIndexByGroupID(ct, tx, -1, relation.OrderIndex+1, groupID)
				if err != nil {
					return err
				}

				_, err = r.groupRepository.UpdateMaxRolloutIndexByGroupID(ct, tx, groupID, -1)
				if err != nil {
					return err
				}
			}
		}

		rollout, err = r.rolloutDao.FindRolloutByIDWithTx(ct, tx, rolloutID)
		if err != nil {
			return err
		}

		rollout.Name = input.Name
		rollout.SelectorID = input.VersionSelectorID
		rollout.ActivatorID = input.ActivatorID
		rollout.IsEnabled = input.IsEnabled
		now := time.Now().UTC()
		rollout.UpdatedAt = &now
		return r.rolloutDao.UpdateRollout(ct, tx, rollout)
	})

	return rollout, err
}

func (r *Rollout) FindGroupRolloutRelationsByGroupID(ct context.Context, groupID uint64) ([]entity.GroupRolloutRelation, *errs.Error) {
	return r.groupRolloutRelationDao.FindGroupRolloutRelationsByGroupID(ct, groupID)
}

func (r *Rollout) GetActiveAppVersionNumberForTeam(ct context.Context, appID uint64, teamID uint64) (*int, *errs.Error) {
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)
	var maxActiveVersionNumber *int
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		groupIDs, err := r.appGroupRelationDao.FindGroupIDsByAppIDWithTx(ct, tx, appID)
		if err != nil {
			return err
		}

		if len(groupIDs) == 0 {
			return nil
		}

		_, err = r.teamDao.FindTeamByIDWithTx(ct, tx, teamID)
		if err != nil {
			if err.Code == errs.NotFound {
				return nil
			}

			return err
		}

		groupIDs, err = r.groupMemberRelationDao.FilterGroupIDsByMemberIDWithTx(ct, tx, groupIDs, teamID)
		if err != nil {
			return err
		}

		for _, teamGroupID := range groupIDs {
			versionNumber, err := r.getTeamGroupActiveVersion(ct, tx, teamID, teamGroupID)
			if err != nil {
				return err
			}

			if versionNumber == nil {
				continue
			}

			if maxActiveVersionNumber == nil || *versionNumber > *maxActiveVersionNumber {
				maxActiveVersionNumber = versionNumber
			}
		}

		return nil
	})

	return maxActiveVersionNumber, err
}

func (r *Rollout) FindRolloutByID(ct context.Context, rolloutID uint64) (entity.Rollout, *errs.Error) {
	rollout, err := r.rolloutDao.FindRolloutByID(ct, rolloutID)
	if err != nil {
		return entity.Rollout{}, err
	}

	return rollout, nil
}

func (r *Rollout) FindActivatorByID(ct context.Context, activatorID uint64) (entity.ActivatorUnion, *errs.Error) {
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)

	var activatorUnion entity.ActivatorUnion
	err := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		activatorUnion, err = r.activatorRepository.FindActivatorByIDWithTx(ct, tx, activatorID)
		return err
	})

	return activatorUnion, err
}

func (r *Rollout) CreateStaticActivator(ct context.Context) (entity.StaticActivator, *errs.Error) {
	genActivatorIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "activatorID"}
	genActivatorRes, rpcErr := r.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genActivatorIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.StaticActivator{}, internalErr
	}

	staticActivator := entity.StaticActivator{
		Activator: entity.Activator{
			ID:        genActivatorRes.UniqueNumber,
			Type:      entity.ActivatorTypeStatic,
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
		return r.activatorRepository.CreateStaticActivator(ct, tx, staticActivator)
	})

	return staticActivator, err
}

func (r *Rollout) CreateTimeRangeActivator(ct context.Context, startAt *time.Time, endAt *time.Time) (entity.TimeRangeActivator, *errs.Error) {
	genActivatorIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "activatorID"}
	genActivatorRes, rpcErr := r.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genActivatorIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.TimeRangeActivator{}, internalErr
	}

	timeRangeActivator := entity.TimeRangeActivator{
		StartAt: startAt,
		EndAt:   endAt,
		Activator: entity.Activator{
			ID:        genActivatorRes.UniqueNumber,
			Type:      entity.ActivatorTypeTimeRange,
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
	genActivatorIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "activatorID"}
	genActivatorRes, rpcErr := r.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genActivatorIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.MaxViewersActivator{}, internalErr
	}
	
	maxViewersActivator := entity.MaxViewersActivator{
		MaxViewers: maxViewers,
		Activator: entity.Activator{
			ID:        genActivatorRes.UniqueNumber,
			Type:      entity.ActivatorTypeMaxViewers,
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
	genActivatorIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "activatorID"}
	genActivatorRes, rpcErr := r.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genActivatorIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.PercentageActivator{}, internalErr
	}

	percentageActivator := entity.PercentageActivator{
		Percentage: percentage,
		Activator: entity.Activator{
			ID:        genActivatorRes.UniqueNumber,
			Type:      entity.ActivatorTypePercentage,
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

func (r *Rollout) UpdateActivator(ct context.Context, activatorID uint64, input UpdateActivatorInput) (entity.ActivatorUnion, *errs.Error) {
	var activatorUnion entity.ActivatorUnion
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)
	now := time.Now().UTC()
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		activator, err := r.activatorDao.FindActivatorByIDWithTx(ct, tx, activatorID)
		if err != nil {
			return err
		}

		updatedActivator := entity.Activator{
			ID:        activatorID,
			Type:      input.Type,
			CreatedAt: activator.CreatedAt,
			UpdatedAt: &now,
		}

		var maxViewers int = 0
		if input.MaxViewers != nil {
			maxViewers = int(*input.MaxViewers)
		}

		var percentage int = 0
		if input.Percentage != nil {
			percentage = int(*input.Percentage)
		}

		if activator.Type != input.Type {
			err := r.activatorRepository.DeletePartialActivator(ct, tx, activatorID)
			if err != nil {
				return err
			}

			createPartialActivatorInput := repository.CreatePartialActivatorInput{
				ID:         activatorID,
				Type:       input.Type,
				StartAt:    input.StartAt,
				EndAt:      input.EndAt,
				MaxViewers: maxViewers,
				Percentage: percentage,
			}
			err = r.activatorRepository.CreatePartialActivator(ct, tx, createPartialActivatorInput)
			if err != nil {
				return err
			}

			err = r.activatorDao.UpdateActivator(ct, tx, updatedActivator)
			if err != nil {
				return err
			}

			activatorUnion, err = r.activatorRepository.GetActivatorUnionFromBaseActivator(ct, tx, updatedActivator)
			return err
		}

		activatorUnion.Type = input.Type
		switch input.Type {
		case entity.ActivatorTypeStatic:
			activatorUnion.StaticActivator = entity.StaticActivator{
				Activator: updatedActivator,
			}
			return r.activatorRepository.UpdateStaticActivator(ct, tx, activatorUnion.StaticActivator)
		case entity.ActivatorTypeTimeRange:
			activatorUnion.TimeRangeActivator = entity.TimeRangeActivator{
				Activator: updatedActivator,
				StartAt:   input.StartAt,
				EndAt:     input.EndAt,
			}
			return r.activatorRepository.UpdateTimeRangeActivator(ct, tx, activatorUnion.TimeRangeActivator)
		case entity.ActivatorTypeMaxViewers:
			activatorUnion.MaxViewersActivator = entity.MaxViewersActivator{
				Activator:  updatedActivator,
				MaxViewers: maxViewers,
			}
			return r.activatorRepository.UpdateMaxViewersActivator(ct, tx, activatorUnion.MaxViewersActivator)
		case entity.ActivatorTypePercentage:
			activatorUnion.PercentageActivator = entity.PercentageActivator{
				Activator:  updatedActivator,
				Percentage: percentage,
			}

			return r.activatorRepository.UpdatePercentageActivator(ct, tx, activatorUnion.PercentageActivator)
		default:
			return errs.NewError(errs.Unknown, fmt.Sprintf("Unknown activator type: %s", input.Type))
		}
	})

	return activatorUnion, err
}

func (r *Rollout) FindVersionSelectorByID(ct context.Context, selectorID uint64) (entity.VersionSelectorUnion, *errs.Error) {
	versionSelectorUnion := entity.VersionSelectorUnion{}
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		versionSelectorUnion, err = r.versionSelectorRepository.FindVersionSelectorByID(ct, tx, selectorID)
		return err
	})

	return versionSelectorUnion, err
}

func (r *Rollout) CreateStaticVersionSelector(ct context.Context, appID uint64, versionNumber int) (entity.StaticVersionSelector, *errs.Error) {
	genSelectorIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "selectorID"}
	genSelectorIDRes, rpcErr := r.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genSelectorIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.StaticVersionSelector{}, internalErr
	}

	staticVersionSelector := entity.StaticVersionSelector{
		VersionSelector: entity.VersionSelector{
			ID:        genSelectorIDRes.UniqueNumber,
			Type:      entity.VersionSelectorTypeStatic,
			CreatedAt: time.Now().UTC(),
		},
		VersionNumber: versionNumber,
	}
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		_, err := r.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
		if err != nil {
			return err
		}

		err = r.versionSelectorRepository.CreateStaticVersionSelector(ct, tx, staticVersionSelector)
		return err
	})

	return staticVersionSelector, err
}

func (r *Rollout) UpdateVersionSelector(ct context.Context, appID uint64, selectorID uint64, input UpdateVersionSelectorInput) (entity.VersionSelectorUnion, *errs.Error) {
	var versionSelectorUnion entity.VersionSelectorUnion
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		versionSelectorUnion, err = r.versionSelectorRepository.FindVersionSelectorByID(ct, tx, selectorID)
		if err != nil {
			return err
		}

		versionSelectorUnion.Type = input.Type
		now := time.Now().UTC()
		switch versionSelectorUnion.Type {
		case entity.VersionSelectorTypeStatic:
			if input.VersionNumber == nil {
				return errs.NewError(errs.InvalidArgument, "Version number is required for static version selector")
			}

			versionNumber := int(*input.VersionNumber)
			_, err := r.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
			if err != nil {
				return err
			}

			versionSelectorUnion.StaticVersionSelector = entity.StaticVersionSelector{
				VersionSelector: entity.VersionSelector{
					ID:        selectorID,
					Type:      entity.VersionSelectorTypeStatic,
					UpdatedAt: &now,
				},
				VersionNumber: versionNumber,
			}
			return r.versionSelectorRepository.UpdateStaticVersionSelector(ct, tx, versionSelectorUnion.StaticVersionSelector)
		case entity.VersionSelectorTypeExperiment:
			for _, versionNumber := range input.VersionNumbers {
				_, err := r.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
				if err != nil {
					return err
				}
			}

			versionSelectorUnion.ExperimentVersionSelector = entity.ExperimentVersionSelector{
				VersionSelector: entity.VersionSelector{
					ID:        selectorID,
					Type:      entity.VersionSelectorTypeExperiment,
					UpdatedAt: &now,
				},
				VersionNumbers: input.VersionNumbers,
			}
			return r.versionSelectorRepository.UpdateExperimentVersionSelector(ct, tx, versionSelectorUnion.ExperimentVersionSelector)
		default:
			return errs.NewError(errs.Unknown, fmt.Sprintf("Unknown version selector type: %s", versionSelectorUnion.Type))
		}
	})

	return versionSelectorUnion, err
}

func (r *Rollout) CreateExperimentVersionSelector(ct context.Context, appID uint64, versionNumbers []int) (entity.ExperimentVersionSelector, *errs.Error) {
	genSelectorIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "selectorID"}
	genSelectorIDRes, rpcErr := r.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genSelectorIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.ExperimentVersionSelector{}, internalErr
	}
	experimentVersionSelector := entity.ExperimentVersionSelector{
		VersionSelector: entity.VersionSelector{
			CreatedAt: time.Now().UTC(),
			ID:        genSelectorIDRes.UniqueNumber,
			Type:      entity.VersionSelectorTypeExperiment,
		},
		VersionNumbers: versionNumbers,
	}
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		for _, versionNumber := range versionNumbers {
			_, err := r.appVersionDao.FindAppVersionByAppIDAndVersionNumber(ct, appID, versionNumber)
			if err != nil {
				return err
			}
		}

		err := r.versionSelectorRepository.CreateExperimentVersionSelector(ct, tx, experimentVersionSelector)
		return err
	})

	return experimentVersionSelector, err
}

func (r *Rollout) DeleteActivator(ct context.Context, activatorID uint64) (entity.ActivatorUnion, *errs.Error) {
	var activatorUnion entity.ActivatorUnion
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)

	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		activatorUnion, err = r.activatorRepository.DeleteActivator(ct, tx, activatorID)
		return err
	})

	return activatorUnion, err
}

func (r *Rollout) DeleteVersionSelector(ct context.Context, selectorID uint64) (entity.VersionSelectorUnion, *errs.Error) {
	var versionSelectorUnion entity.VersionSelectorUnion
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		versionSelectorUnion, err = r.versionSelectorRepository.FindVersionSelectorByID(ct, tx, selectorID)
		if err != nil {
			return err
		}

		return r.versionSelectorRepository.DeleteVersionSelector(ct, tx, selectorID)
	})

	return versionSelectorUnion, err
}

func (r *Rollout) getTeamGroupActiveVersion(
	ct context.Context,
	tx *cloudTransaction.Transaction,
	teamID uint64,
	groupID uint64,
) (*int, *errs.Error) {
	groupRolloutRelations, err := r.groupRolloutRelationDao.FindGroupRolloutRelationsByGroupIDWithTx(ct, tx, groupID)
	if err != nil {
		return nil, err
	}

	sort.Slice(groupRolloutRelations, func(i, j int) bool {
		return groupRolloutRelations[i].OrderIndex < groupRolloutRelations[j].OrderIndex
	})
	sortedRolloutIDs := collect.Map(groupRolloutRelations, func(groupRolloutRelation entity.GroupRolloutRelation, index int) uint64 {
		return groupRolloutRelation.RolloutID
	})
	rawRollouts, err := r.rolloutDao.FindRolloutsByIDsWithTx(ct, tx, sortedRolloutIDs)
	if err != nil {
		return nil, err
	}

	rollouts := make([]rollout.Rollout, 0)
	for _, rawRollout := range rawRollouts {
		rollout, err := r.newRollout(ct, tx, rawRollout)
		if err != nil {
			return nil, err
		}

		rollouts = append(rollouts, rollout)
	}

	orderedRollouts := rollout.OrderedRollouts(rollouts)
	versionNumber, err := orderedRollouts.GetVersionNumber(ct, teamID)
	return versionNumber, err
}

func (r *Rollout) getRolloutVersionSelector(
	ct context.Context,
	tx *cloudTransaction.Transaction,
	rolloutID uint64,
	selectorID uint64,
) (rollout.VersionSelector, *errs.Error) {
	var versionSelector rollout.VersionSelector

	rawVersionSelector, err := r.versionSelectorRepository.FindVersionSelectorByID(ct, tx, selectorID)
	if err != nil {
		return nil, err
	}

	switch rawVersionSelector.Type {
	case entity.VersionSelectorTypeStatic:
		versionSelector = rollout.NewStaticVersionSelector(rawVersionSelector.StaticVersionSelector.VersionNumber)
	case entity.VersionSelectorTypeExperiment:
		store := store.NewExperimentVersionSelector(r.logger, r.transactionFactory, r.stateSyncer, r.rolloutViewerDao, rolloutID)
		versionSelector = rollout.NewExperimentVersionSelector(store, randgen.NewBuiltinRanGen(), rawVersionSelector.ExperimentVersionSelector.VersionNumbers)
	default:
		return nil, errs.NewError(errs.Unknown, fmt.Sprintf("unknown version selector type %s", rawVersionSelector.Type))
	}

	return versionSelector, nil
}

func (r *Rollout) getRolloutActivator(
	ct context.Context,
	tx *cloudTransaction.Transaction,
	rolloutID uint64,
	activatorID uint64,
) (rollout.Activator, *errs.Error) {
	var activator rollout.Activator
	activatorUnion, err := r.activatorRepository.FindActivatorByIDWithTx(ct, tx, activatorID)
	if err != nil {
		return nil, err
	}

	switch activatorUnion.Type {
	case entity.ActivatorTypeStatic:
		activator = rollout.NewStaticActivator()
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

func (r *Rollout) newRollout(
	ct context.Context,
	tx *cloudTransaction.Transaction,
	rawRollout entity.Rollout,
) (rollout.Rollout, *errs.Error) {
	activatorID := rawRollout.ActivatorID
	versionSelector, err := r.getRolloutVersionSelector(ct, tx, rawRollout.ID, rawRollout.SelectorID)
	if err != nil {
		return rollout.Rollout{}, err
	}

	activator, err := r.getRolloutActivator(ct, tx, rawRollout.ID, activatorID)
	if err != nil {
		return rollout.Rollout{}, err
	}

	rolloutStore := store.NewRollout(r.logger, r.transactionFactory, r.stateSyncer, r.rolloutDao, rawRollout.ID)
	return rollout.NewRollout(ct, rolloutStore, activator, versionSelector)
}

func NewRollout(
	logger telemetry.Logger,
	cloudClientRegistry *client.Registry,
	transactionFactory cloudTransaction.Factory,
	stateSyncer *realtime.StateSyncer,
	appGroupRelationDao dao.AppGroupRelation,
	rolloutDao dao.Rollout,
	rolloutViewerDao dao.RolloutViewer,
	groupRolloutRelationDao dao.GroupRolloutRelation,
	appRolloutRelationDao dao.AppRolloutRelation,
	groupMemberRelationDao dao.GroupMemberRelation,
	appVersionDao dao.AppVersion,
	versionSelectorRepository *repository.VersionSelector,
	activatorRepository *repository.Activator,
	groupRepository *repository.Group,
	activatorDao dao.Activator,
	teamDao dao.Team,
) *Rollout {
	return &Rollout{
		logger:                    logger,
		cloudClientRegistry:       cloudClientRegistry,
		transactionFactory:        transactionFactory,
		stateSyncer:               stateSyncer,
		appGroupRelationDao:       appGroupRelationDao,
		rolloutDao:                rolloutDao,
		rolloutViewerDao:          rolloutViewerDao,
		groupRolloutRelationDao:   groupRolloutRelationDao,
		appRolloutRelationDao:     appRolloutRelationDao,
		groupMemberRelationDao:    groupMemberRelationDao,
		appVersionDao:             appVersionDao,
		versionSelectorRepository: versionSelectorRepository,
		activatorRepository:       activatorRepository,
		groupRepository:           groupRepository,
		activatorDao:              activatorDao,
		teamDao:                   teamDao,
	}
}
