package daotestv2

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintParticipant struct {
	db *dbtest.InMemoryDB
}

var _ daov2.SprintParticipant = (*SprintParticipant)(nil)

func (s SprintParticipant) FindParticipantIDsBySprintID(ct context.Context, tx *transaction.Transaction, sprintID uint64) ([]uint64, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s SprintParticipant) FindParticipantsBySprintID(ct context.Context, tx *transaction.Transaction, sprintID uint64) ([]entity.SprintParticipant, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s SprintParticipant) FindParticipant(ct context.Context, tx *transaction.Transaction, sprintID uint64, participantUserID uint64) (entity.SprintParticipant, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s SprintParticipant) CreateSprintParticipant(ct context.Context, tx *transaction.Transaction, participant entity.SprintParticipant) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (s SprintParticipant) UpdateSprintParticipant(ct context.Context, tx *transaction.Transaction, participant entity.SprintParticipant) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (s SprintParticipant) DeleteSprintParticipant(ct context.Context, tx *transaction.Transaction, sprintID uint64, userID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewSprintParticipant(db *dbtest.InMemoryDB) SprintParticipant {
	return SprintParticipant{db: db}
}
