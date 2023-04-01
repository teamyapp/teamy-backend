package daov2

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamMember interface {
	FindTeamIDsByUserID(ct context.Context, userID uint64) ([]uint64, *errs.Error)
	FindTeamMembersByTeamID(ct context.Context, teamID uint64) ([]entity.TeamMember, *errs.Error)
	FindTeamIDsByUserIDWithTx(ct context.Context, tx *transaction.Transaction, userID uint64) ([]uint64, *errs.Error)
	FindTeamMemberIDsByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]uint64, *errs.Error)
	FindTeamMembersByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.TeamMember, *errs.Error)
	FindTeamMemberWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64, userID uint64) (entity.TeamMember, *errs.Error)
	CreateTeamMember(ct context.Context, tx *transaction.Transaction, teamMember entity.TeamMember) *errs.Error
	UpdateTeamMember(ct context.Context, tx *transaction.Transaction, teamMember entity.TeamMember) *errs.Error
	DeleteTeamMember(ct context.Context, tx *transaction.Transaction, teamID uint64, userID uint64) *errs.Error
}
