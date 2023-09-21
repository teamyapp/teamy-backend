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

func (v *VersionSelector) CreateVersionSelector(
	ct context.Context,
	tx *transaction.Transaction,
	versionSelectorType entity.VersionSelectorType,
	versionNumbers []int,
) *errs.Error {
	genSelectorIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "selectorID"}
	genSelectorIDRes, rpcErr := v.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genSelectorIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return internalErr
	}

	versionSelector := entity.VersionSelector{
		ID:   genSelectorIDRes.UniqueNumber,
		Type: versionSelectorType,
	}

	err := v.versionSelectorDao.CreateVersionSelector(ct, tx, versionSelector)
	if err != nil {
		return err
	}

	for _, versionNumber := range versionNumbers {
		err := v.versionSelectorVersionRelationDao.CreateVersionSelectorVersionRelation(ct, tx, entity.VersionSelectorVersionRelation{
			VersionSelectorID: versionSelector.ID,
			VersionNumber:     versionNumber,
		})
		if err != nil {
			return err
		}
	}

	return nil
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
		versionNumber, err := v.FindVersionNumberByStaticVersionSelectorID(ct, tx, versionSelectorID)
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
		versionNumbers, err := v.FindVersionNumbersByExperimentVersionSelectorID(ct, tx, versionSelectorID)
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

func (v *VersionSelector) FindVersionNumbersByExperimentVersionSelectorID(ct context.Context, tx *transaction.Transaction, versionSelectorID uint64) ([]int, *errs.Error) {
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

func (v *VersionSelector) FindVersionNumberByStaticVersionSelectorID(ct context.Context, tx *transaction.Transaction, versionSelectorID uint64) (int, *errs.Error) {
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

func NewVersionSelector(logger telemetry.Logger, versionSelectorDao dao.VersionSelector) *VersionSelector {
	return &VersionSelector{
		logger:             logger,
		versionSelectorDao: versionSelectorDao,
	}
}
