package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppRolloutRelation struct {
}

var _ dao.AppRolloutRelation = (*AppRolloutRelation)(nil)

func (*AppRolloutRelation) FindRolloutIDsByAppIDAndRelationType(ct context.Context, appID uint64, rolloutType entity.AppRolloutRelationType) ([]uint64, *errs.Error) {
	panic("unimplemented")
}

func (*AppRolloutRelation) CreateAppRolloutRelation(ct context.Context, appRolloutRelation entity.AppRolloutRelation) (entity.AppRolloutRelation, *errs.Error) {
	panic("unimplemented")
}

func NewAppRolloutRelation() *AppRolloutRelation {
	return &AppRolloutRelation{}
}
