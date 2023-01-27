package dao

import (
	"context"

	"github.com/teamyapp/teamy-backend/core/entity"
)

type Message interface {
	FindMessageByID(ct context.Context, messageID uint64) (entity.Message, error)
	FindMessagesByThreadID(ct context.Context, threadID uint64) ([]entity.Message, error)
	CreateMessage(ct context.Context, message entity.Message) error
	UpdateMessage(ct context.Context, message entity.Message) error
	DeleteMessage(ct context.Context, messageID uint64) error
}
