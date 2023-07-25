package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamGroup interface {
	FindGroupByTeamIDAndLabel(
		ct context.Context,
		teamID uint64,
		groupLabel string,
	) (entity.TeamGroup, *errs.Error)
	FindGroupByTeamIDAndLabelWithTx(
		ct context.Context,
		tx *transaction.Transaction,
		teamID uint64,
		groupLabel string,
	) (entity.TeamGroup, *errs.Error)
	CreateGroup(
		ct context.Context,
		tx *transaction.Transaction,
		group entity.TeamGroup,
	) *errs.Error
	DeleteGroup(
		ct context.Context,
		tx *transaction.Transaction,
		teamID uint64,
		groupLabel string,
	) *errs.Error
}
