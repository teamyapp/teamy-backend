package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamGroupRelation struct{}

// CreateTeamGroupRelation implements dao.TeamGroupRelation.
func (*TeamGroupRelation) CreateTeamGroupRelation(ct context.Context, teamGroupRelation entity.TeamGroupRelation) (entity.TeamGroupRelation, *errs.Error) {
	panic("unimplemented")
}

// DeleteTeamGroupRelation implements dao.TeamGroupRelation.
func (*TeamGroupRelation) DeleteTeamGroupRelation(ct context.Context, teamID uint64, groupID uint64) *errs.Error {
	panic("unimplemented")
}

// FindTeamIDsByGroupID implements dao.TeamGroupRelation.
func (*TeamGroupRelation) FindTeamIDsByGroupID(ct context.Context, groupID uint64) ([]uint64, *errs.Error) {
	panic("unimplemented")
}

var _ dao.TeamGroupRelation = (*TeamGroupRelation)(nil)

func NewTeamGroupRelation() *TeamGroupRelation {
	return &TeamGroupRelation{}
}
