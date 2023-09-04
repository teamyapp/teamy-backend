package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamGroupRelation interface {
	FindTeamIDsByGroupID(ct context.Context, groupID uint64) ([]uint64, *errs.Error)
	CreateTeamGroupRelation(ct context.Context, teamGroupRelation entity.TeamGroupRelation) (entity.TeamGroupRelation, *errs.Error)
	DeleteTeamGroupRelation(ct context.Context, teamID uint64, groupID uint64) *errs.Error
}
