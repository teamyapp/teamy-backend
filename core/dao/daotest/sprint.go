package daotest

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Sprint struct {
	db *dbtest.InMemoryDB
}

var _ dao.Sprint = (*Sprint)(nil)

func (s Sprint) FindSprintByID(ct context.Context, sprintID uint64) (entity.Sprint, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s Sprint) FindSprintsByIDs(ct context.Context, sprintIDs []uint64) ([]entity.Sprint, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s Sprint) FindSprintsByTeamID(ct context.Context, teamID uint64) ([]entity.Sprint, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s Sprint) FindAllSprints(ct context.Context) ([]entity.Sprint, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s Sprint) CreateSprint(ct context.Context, sprint entity.Sprint) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (s Sprint) DeleteSprint(ct context.Context, sprintID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewSprint(db *dbtest.InMemoryDB) Sprint {
	return Sprint{db: db}
}
