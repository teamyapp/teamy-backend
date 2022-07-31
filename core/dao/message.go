package dao

import (
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Message interface {
	FindMessageByID(messageID uint64) (entity.Message, error)
	FindMessagesByThreadID(threadID uint64) ([]entity.Message, error)
	CreateMessage(message entity.Message) error
	UpdateMessage(message entity.Message) error
	DeleteMessage(messageID uint64) error
}
