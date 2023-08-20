package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Sprint interface {
	FindSprintByID(ct context.Context, sprintID uint64) (entity.Sprint, *errs.Error)
	FindSprintsByTeamID(ct context.Context, teamID uint64) ([]entity.Sprint, *errs.Error)
	FindAllSprints(ct context.Context) ([]entity.Sprint, *errs.Error)
	FindSprintByIDWithTx(ct context.Context, tx *transaction.Transaction, sprintID uint64) (entity.Sprint, *errs.Error)
	FindSprintsByIDsWithTx(ct context.Context, tx *transaction.Transaction, sprintIDs []uint64) ([]entity.Sprint, *errs.Error)
	FindSprintsByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.Sprint, *errs.Error)
	FindAllSprintsWithTx(ct context.Context, tx *transaction.Transaction) ([]entity.Sprint, *errs.Error)
	CreateSprint(ct context.Context, tx *transaction.Transaction, sprint entity.Sprint) *errs.Error
	DeleteSprint(ct context.Context, tx *transaction.Transaction, sprintID uint64) *errs.Error
}
