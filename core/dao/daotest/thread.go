package daotest

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
)

type Thread struct {
	db *dbtest.InMemoryDB
}

var _ dao.Thread = (*Thread)(nil)

func (t Thread) CreateThread(ct context.Context, threadID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (t Thread) DeleteThread(ct context.Context, threadID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewThread(db *dbtest.InMemoryDB) Thread {
	return Thread{db: db}
}
