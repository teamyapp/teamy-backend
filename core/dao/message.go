package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Message interface {
	FindMessageByID(ct context.Context, messageID uint64) (entity.Message, *errs.Error)
	FindMessagesByThreadID(ct context.Context, threadID uint64) ([]entity.Message, *errs.Error)
	FindMessageByIDWithTx(ct context.Context, tx *transaction.Transaction, messageID uint64) (entity.Message, *errs.Error)
	FindMessagesByThreadIDWithTx(ct context.Context, tx *transaction.Transaction, threadID uint64) ([]entity.Message, *errs.Error)
	CreateMessage(ct context.Context, tx *transaction.Transaction, message entity.Message) *errs.Error
	UpdateMessage(ct context.Context, tx *transaction.Transaction, message entity.Message) *errs.Error
	DeleteMessage(ct context.Context, tx *transaction.Transaction, messageID uint64) *errs.Error
}
