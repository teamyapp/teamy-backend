package daotest

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintParticipant struct {
	db *dbtest.InMemoryDB
}

var _ dao.SprintParticipant = (*SprintParticipant)(nil)

func (s SprintParticipant) FindParticipantIDsBySprintID(ct context.Context, sprintID uint64) ([]uint64, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s SprintParticipant) FindParticipantsBySprintID(ct context.Context, sprintID uint64) ([]entity.SprintParticipant, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s SprintParticipant) FindParticipant(ct context.Context, sprintID uint64, participantUserID uint64) (entity.SprintParticipant, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s SprintParticipant) CreateSprintParticipant(ct context.Context, participant entity.SprintParticipant) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (s SprintParticipant) UpdateSprintParticipant(ct context.Context, participant entity.SprintParticipant) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (s SprintParticipant) DeleteSprintParticipant(ct context.Context, sprintID uint64, userID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewSprintParticipant(db *dbtest.InMemoryDB) SprintParticipant {
	return SprintParticipant{db: db}
}
