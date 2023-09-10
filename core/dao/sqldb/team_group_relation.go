package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamGroupRelation struct{}

var _ dao.TeamGroupRelation = (*TeamGroupRelation)(nil)

func (*TeamGroupRelation) FindTeamIDsByGroupID(ct context.Context, groupID uint64) ([]uint64, *errs.Error) {
	panic("unimplemented")
}
func (*TeamGroupRelation) CreateTeamGroupRelation(ct context.Context, teamGroupRelation entity.TeamGroupRelation) (entity.TeamGroupRelation, *errs.Error) {
	panic("unimplemented")
}

func (*TeamGroupRelation) DeleteTeamGroupRelation(ct context.Context, teamID uint64, groupID uint64) *errs.Error {
	panic("unimplemented")
}

func NewTeamGroupRelation() *TeamGroupRelation {
	return &TeamGroupRelation{}
}
