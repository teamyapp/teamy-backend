package daotestv2

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Sprint struct {
	db *dbtest.InMemoryDB
}

var _ daov2.Sprint = (*Sprint)(nil)

func (s Sprint) FindSprintByID(ct context.Context, tx *transaction.Transaction, sprintID uint64) (entity.Sprint, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s Sprint) FindSprintsByIDs(ct context.Context, tx *transaction.Transaction, sprintIDs []uint64) ([]entity.Sprint, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s Sprint) FindSprintsByTeamID(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.Sprint, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s Sprint) FindAllSprints(ct context.Context, tx *transaction.Transaction) ([]entity.Sprint, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s Sprint) CreateSprint(ct context.Context, tx *transaction.Transaction, sprint entity.Sprint) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (s Sprint) DeleteSprint(ct context.Context, tx *transaction.Transaction, sprintID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewSprint(db *dbtest.InMemoryDB) Sprint {
	return Sprint{db: db}
}
