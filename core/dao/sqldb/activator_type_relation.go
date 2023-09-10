package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type ActivatorTypeRelation struct {
}

var _ dao.ActivatorTypeRelation = (*ActivatorTypeRelation)(nil)

func (*ActivatorTypeRelation) FindActivatorTypeByID(ct context.Context, activatorID uint64) (entity.ActivatorType, *errs.Error) {
	panic("unimplemented")
}

func (*ActivatorTypeRelation) CreateActivatorTypeRelation(ct context.Context, ActivatorTypeRelation entity.ActivatorTypeRelation) (entity.ActivatorType, *errs.Error) {
	panic("unimplemented")
}

func NewActivatorTypeRelation() *ActivatorTypeRelation {
	return &ActivatorTypeRelation{}
}
