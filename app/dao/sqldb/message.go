package sqldb

import (
	"database/sql"

	"github.com/teamyapp/teamy-backend/app/dao"
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type Message struct {
	db *sql.DB
}

var _ dao.Message = (*Message)(nil)

func (m Message) FindMessageByID(messageID uint64) (entityv2.Message, error) {
	//TODO implement me
	panic("implement me")
}

func (m Message) FindMessagesByIDs(messageIDs []uint64) ([]entityv2.Message, error) {
	//TODO implement me
	panic("implement me")
}

func (m Message) FindMessagesByThreadID(threadID uint64) ([]entityv2.Message, error) {
	//TODO implement me
	panic("implement me")
}

func NewMessage(sqlDB *sql.DB) Message {
	return Message{db: sqlDB}
}
