package daotest

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Message struct {
	db *dbtest.InMemoryDB
}

var _ dao.Message = (*Message)(nil)

func (m Message) FindMessageByID(ct context.Context, messageID uint64) (entity.Message, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (m Message) FindMessagesByThreadID(ct context.Context, threadID uint64) ([]entity.Message, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (m Message) CreateMessage(ct context.Context, message entity.Message) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (m Message) UpdateMessage(ct context.Context, message entity.Message) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (m Message) DeleteMessage(ct context.Context, messageID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewMessage(db *dbtest.InMemoryDB) Message {
	return Message{db: db}
}
