package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamMemberGroupInvitationRelation interface {
	FindInvitationIDsByTeamMemberGroupID(
		ct context.Context,
		tx *transaction.Transaction,
		teamMemberGroupID uint64,
	) ([]uint64, *errs.Error)
	CreateTeamMemberGroupInvitationRelation(
		ct context.Context,
		tx *transaction.Transaction,
		relation entity.TeamMemberGroupInvitationRelation,
	) *errs.Error
	DeleteTeamMemberGroupInvitationRelation(
		ct context.Context,
		tx *transaction.Transaction,
		relation entity.TeamMemberGroupInvitationRelation,
	) *errs.Error
	DeleteTeamMemberGroupInvitationRelationsByGroupID(
		ct context.Context,
		tx *transaction.Transaction,
		groupID uint64,
	) *errs.Error
}
