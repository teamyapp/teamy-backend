package daov2

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Sprint interface {
	FindSprintByID(ct context.Context, tx *transaction.Transaction, sprintID uint64) (entity.Sprint, *errs.Error)
	FindSprintsByIDs(ct context.Context, tx *transaction.Transaction, sprintIDs []uint64) ([]entity.Sprint, *errs.Error)
	FindSprintsByTeamID(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.Sprint, *errs.Error)
	FindAllSprints(ct context.Context, tx *transaction.Transaction) ([]entity.Sprint, *errs.Error)
	CreateSprint(ct context.Context, tx *transaction.Transaction, sprint entity.Sprint) *errs.Error
	DeleteSprint(ct context.Context, tx *transaction.Transaction, sprintID uint64) *errs.Error
}
