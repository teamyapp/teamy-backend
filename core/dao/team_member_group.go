package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao/entity"
)

type TeamMemberGroup interface {
	FindMemberGroupByID(
		ct context.Context,
		tx *transaction.Transaction,
		id uint64,
	) (entity.TeamMemberGroup, *errs.Error)
	FindMemberGroupsByTeamID(
		ct context.Context,
		tx *transaction.Transaction,
		teamID uint64,
	) ([]entity.TeamMemberGroup, *errs.Error)
	FindMaxTeamMemberGroupOrderIndexByTeamID(
		ct context.Context,
		tx *transaction.Transaction,
		teamID uint64,
	) (int, *errs.Error)
	CreateMemberGroup(
		ct context.Context,
		tx *transaction.Transaction,
		memberGroup entity.TeamMemberGroup,
	) *errs.Error
	UpdateMemberGroup(
		ct context.Context,
		tx *transaction.Transaction,
		memberGroup entity.TeamMemberGroup,
	) *errs.Error
	DeleteMemberGroup(
		ct context.Context,
		tx *transaction.Transaction,
		id uint64,
	) *errs.Error
}
