package daotestv2

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintTaskRelation struct {
	db *dbtest.InMemoryDB
}

var _ daov2.SprintTaskRelation = (*SprintTaskRelation)(nil)

func (s SprintTaskRelation) FindTaskIDsBySprintIDWithTx(ct context.Context, tx *transaction.Transaction, sprintID uint64) ([]uint64, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s SprintTaskRelation) FindSprintIDsByTaskIDWithTx(ct context.Context, tx *transaction.Transaction, taskID uint64) ([]uint64, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s SprintTaskRelation) CreateSprintTaskRelation(ct context.Context, tx *transaction.Transaction, relation entity.SprintTaskRelation) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (s SprintTaskRelation) DeleteSprintTaskRelation(ct context.Context, tx *transaction.Transaction, sprintID uint64, taskID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewSprintTaskRelation(db *dbtest.InMemoryDB) SprintTaskRelation {
	return SprintTaskRelation{db: db}
}
