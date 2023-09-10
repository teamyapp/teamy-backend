package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type GroupRolloutRelation struct{}

// FindGroupRolloutRelationsByGroupID implements dao.GroupRolloutRelation.
func (*GroupRolloutRelation) FindGroupRolloutRelationsByGroupID(ct context.Context, groupID uint64) ([]entity.GroupRolloutRelation, *errs.Error) {
	panic("unimplemented")
}

// FindRolloutIDsByGroupID implements dao.GroupRolloutRelation.
func (*GroupRolloutRelation) FindRolloutIDsByGroupID(ct context.Context, groupID uint64) ([]uint64, *errs.Error) {
	panic("unimplemented")
}

var _ dao.GroupRolloutRelation = (*GroupRolloutRelation)(nil)

func NewGroupRolloutRelation() *GroupRolloutRelation {
	return &GroupRolloutRelation{}
}
