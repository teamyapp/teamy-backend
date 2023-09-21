package repository

import (
	"context"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type VersionSelector struct {
	logger                            telemetry.Logger
	cloudClientRegistry               *client.Registry
	versionSelectorDao                dao.VersionSelector
	versionSelectorVersionRelationDao dao.VersionSelectorVersionRelation
}

func (v *VersionSelector) CreateStaticVersionSelector(
	ct context.Context,
	tx *transaction.Transaction,
	versionNumber int,
) (entity.StaticVersionSelector, *errs.Error) {
	genSelectorIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "selectorID"}
	genSelectorIDRes, rpcErr := v.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genSelectorIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.StaticVersionSelector{}, internalErr
	}
	versionSelector := entity.VersionSelector{
		ID:   genSelectorIDRes.UniqueNumber,
		Type: entity.VersionSelectorTypeStatic,
	}

	err := v.versionSelectorDao.CreateVersionSelector(ct, tx, versionSelector)
	if err != nil {
		return entity.StaticVersionSelector{}, err
	}

	err = v.versionSelectorVersionRelationDao.CreateVersionSelectorVersionRelation(ct, tx, entity.VersionSelectorVersionRelation{
		VersionSelectorID: versionSelector.ID,
		VersionNumber:     versionNumber,
	})
	if err != nil {
		return entity.StaticVersionSelector{}, err
	}

	return entity.StaticVersionSelector{
		VersionSelector: versionSelector,
		VersionNumber:   versionNumber,
	}, nil
}

func (v *VersionSelector) CreateExperimentVersionSelector(
	ct context.Context,
	tx *transaction.Transaction,
	versionNumbers []int,
) (entity.ExperimentVersionSelector, *errs.Error) {
	genSelectorIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "selectorID"}
	genSelectorIDRes, rpcErr := v.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genSelectorIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.ExperimentVersionSelector{}, internalErr
	}
	versionSelector := entity.VersionSelector{
		ID:   genSelectorIDRes.UniqueNumber,
		Type: entity.VersionSelectorTypeExperiment,
	}

	err := v.versionSelectorDao.CreateVersionSelector(ct, tx, versionSelector)
	if err != nil {
		return entity.ExperimentVersionSelector{}, err
	}

	for _, versionNumber := range versionNumbers {
		err = v.versionSelectorVersionRelationDao.CreateVersionSelectorVersionRelation(ct, tx, entity.VersionSelectorVersionRelation{
			VersionSelectorID: versionSelector.ID,
			VersionNumber:     versionNumber,
		})
		if err != nil {
			return entity.ExperimentVersionSelector{}, err
		}
	}

	return entity.ExperimentVersionSelector{
		VersionSelector: versionSelector,
		VersionNumbers:  versionNumbers,
	}, nil
}

func (v *VersionSelector) FindVersionSelectorByID(
	ct context.Context,
	tx *transaction.Transaction,
	versionSelectorID uint64,
) (entity.VersionSelectorUnion, *errs.Error) {
	versionSelector, err := v.versionSelectorDao.FindVersionSelectorByIDWithTx(ct, tx, versionSelectorID)
	if err != nil {
		return entity.VersionSelectorUnion{}, err
	}

	switch versionSelector.Type {
	case entity.VersionSelectorTypeStatic:
		versionNumber, err := v.findVersionNumberByStaticVersionSelectorID(ct, tx, versionSelectorID)
		if err != nil {
			return entity.VersionSelectorUnion{}, err
		}

		return entity.VersionSelectorUnion{
			Type: entity.VersionSelectorTypeStatic,
			StaticVersionSelector: entity.StaticVersionSelector{
				VersionSelector: versionSelector,
				VersionNumber:   versionNumber,
			},
		}, nil
	case entity.VersionSelectorTypeExperiment:
		versionNumbers, err := v.findVersionNumbersByExperimentVersionSelectorID(ct, tx, versionSelectorID)
		if err != nil {
			return entity.VersionSelectorUnion{}, err
		}

		return entity.VersionSelectorUnion{
			Type: entity.VersionSelectorTypeExperiment,
			ExperimentVersionSelector: entity.ExperimentVersionSelector{
				VersionSelector: versionSelector,
				VersionNumbers:  versionNumbers,
			},
		}, nil
	default:
		return entity.VersionSelectorUnion{}, errs.NewError(errs.InvalidArgument, "invalid version selector type")
	}
}

func (v *VersionSelector) findVersionNumberByStaticVersionSelectorID(ct context.Context, tx *transaction.Transaction, versionSelectorID uint64) (int, *errs.Error) {
	versionSelector, err := v.versionSelectorDao.FindVersionSelectorByIDWithTx(ct, tx, versionSelectorID)
	if err != nil {
		return 0, err
	}

	if versionSelector.Type != entity.VersionSelectorTypeStatic {
		return 0, errs.NewError(errs.InvalidArgument, "version selector is not a static")
	}

	versionNumbers, err := v.versionSelectorVersionRelationDao.FindVersionNumbersBySelectorID(ct, versionSelectorID)
	if err != nil {
		return 0, err
	}

	if len(versionNumbers) == 0 {
		return 0, errs.NewError(errs.NotFound, "app version not found")
	}

	if len(versionNumbers) != 1 {
		return 0, errs.NewError(errs.InvalidArgument, "static version selector has more than one version number")
	}

	return versionNumbers[0], nil
}

func (v *VersionSelector) findVersionNumbersByExperimentVersionSelectorID(ct context.Context, tx *transaction.Transaction, versionSelectorID uint64) ([]int, *errs.Error) {
	versionSelector, err := v.versionSelectorDao.FindVersionSelectorByIDWithTx(ct, tx, versionSelectorID)
	if err != nil {
		return nil, err
	}

	if versionSelector.Type != entity.VersionSelectorTypeExperiment {
		return nil, errs.NewError(errs.InvalidArgument, "version selector is not an experiment")
	}

	versionNumbers, err := v.versionSelectorVersionRelationDao.FindVersionNumbersBySelectorID(ct, versionSelectorID)
	if err != nil {
		return nil, err
	}

	if len(versionNumbers) == 0 {
		return nil, errs.NewError(errs.NotFound, "app version not found")
	}

	return versionNumbers, nil
}

func NewVersionSelector(logger telemetry.Logger, versionSelectorDao dao.VersionSelector) *VersionSelector {
	return &VersionSelector{
		logger:             logger,
		versionSelectorDao: versionSelectorDao,
	}
}
