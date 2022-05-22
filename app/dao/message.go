package dao

import "github.com/teamyapp/teamy-backend/app/entityv2"

type Message interface {
	FindMessageByID(messageID uint64) (entityv2.Message, error)
	FindMessagesByIDs(messageIDs []uint64) ([]entityv2.Message, error)
	FindMessagesByThreadID(threadID uint64) ([]entityv2.Message, error)
	CreateMessage(message entityv2.Message) error
	UpdateMessage(message entityv2.Message) error
	DeleteMessage(messageID uint64) error
}
