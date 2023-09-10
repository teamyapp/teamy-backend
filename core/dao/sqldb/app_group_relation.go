package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppGroupRelation struct {
}

func (*AppGroupRelation) CreateAppGroupRelation(ct context.Context, appGroupRelation entity.AppGroupRelation) (entity.AppGroupRelation, *errs.Error) {
	panic("unimplemented")
}

func (*AppGroupRelation) FindAppIDByGroupID(ct context.Context, groupID uint64) (uint64, *errs.Error) {
	panic("unimplemented")
}

func (*AppGroupRelation) FindGroupIDsByAppIDAndRelationType(ct context.Context, appID uint64, appGroupRelationType entity.AppGroupRelationType) ([]uint64, *errs.Error) {
	panic("unimplemented")
}

var _ dao.AppGroupRelation = (*AppGroupRelation)(nil)

func NewAppGroupRelation() *AppGroupRelation {
	return &AppGroupRelation{}
}
