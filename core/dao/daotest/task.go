package daotest

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Task struct {
	db *dbtest.InMemoryDB
}

var _ dao.Task = (*Task)(nil)

func (t Task) FindTaskByID(ct context.Context, taskID uint64) (entity.Task, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t Task) FindTaskByCommentsThreadID(ct context.Context, commentThreadID uint64) (entity.Task, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t Task) FindTasksByIDs(ct context.Context, taskIDs []uint64) ([]entity.Task, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t Task) FindTasksByTeamID(ct context.Context, teamID uint64) ([]entity.Task, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t Task) FindAllTasks(ct context.Context) ([]entity.Task, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t Task) CreateTask(ct context.Context, task entity.Task) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (t Task) UpdateTask(ct context.Context, task entity.Task) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (t Task) DeleteTask(ct context.Context, taskID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewTask(db *dbtest.InMemoryDB) Task {
	return Task{db: db}
}
