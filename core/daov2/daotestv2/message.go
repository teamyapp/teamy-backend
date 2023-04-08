package daotestv2

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Message struct {
	db                 *dbtest.InMemoryDB
	transactionFactory transaction.Factory
}

var _ daov2.Message = (*Message)(nil)

func (m Message) FindMessageByID(ct context.Context, messageID uint64) (entity.Message, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := m.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.Message{}, err
	}

	defer tx.Rollback()
	return m.FindMessageByIDWithTx(ct, tx, messageID)
}

func (m Message) FindMessagesByThreadID(ct context.Context, threadID uint64) ([]entity.Message, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := m.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return m.FindMessagesByThreadIDWithTx(ct, tx, threadID)
}

func (m Message) FindMessageByIDWithTx(ct context.Context, tx *transaction.Transaction, messageID uint64) (entity.Message, *errs.Error) {
	var message entity.Message
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := m.db.GetTable(MessageTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currMessage := rawRow.(entity.Message)
				if currMessage.ID == messageID {
					message = currMessage
					return nil
				}
			}

			return errs.NewError(errs.NotFound, fmt.Sprintf("row not found: messageID=%v", messageID))
		},
	})
	return message, err
}

func (m Message) FindMessagesByThreadIDWithTx(ct context.Context, tx *transaction.Transaction, threadID uint64) ([]entity.Message, *errs.Error) {
	var messages []entity.Message
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := m.db.GetTable(MessageTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currMessage := rawRow.(entity.Message)
				if currMessage.ThreadID == threadID {
					messages = append(messages, currMessage)
				}
			}

			return nil
		},
	})
	return messages, err
}

func (m Message) CreateMessage(ct context.Context, tx *transaction.Transaction, message entity.Message) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := m.db.GetTable(MessageTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				currMessage := row.(entity.Message)
				if currMessage.ID == message.ID {
					return errs.NewError(errs.Unknown, fmt.Sprintf("row already exist: messageID=%v", message.ID))
				}
			}

			table.Rows = append(table.Rows, message)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := m.db.GetTable(MessageTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currMessage := row.(entity.Message)
				if currMessage.ID == message.ID {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (m Message) UpdateMessage(ct context.Context, tx *transaction.Transaction, message entity.Message) *errs.Error {
	oldMessage, internalErr := m.FindMessageByIDWithTx(ct, tx, message.ID)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := m.db.GetTable(MessageTableName)
			if err != nil {
				return err
			}

			for i, row := range table.Rows {
				currMessage := row.(entity.Message)
				if currMessage.ID == message.ID {
					table.Rows[i] = message
					return nil
				}
			}

			return errs.NewError(errs.Unknown, fmt.Sprintf("row not exist: messageID=%v", message.ID))
		},
		Undo: func() *errs.Error {
			table, err := m.db.GetTable(MessageTableName)
			if err != nil {
				return err
			}

			for index, row := range table.Rows {
				currMessage := row.(entity.Message)
				if currMessage.ID == message.ID {
					table.Rows[index] = oldMessage
				}
			}

			return nil
		},
	})
}

func (m Message) DeleteMessage(ct context.Context, tx *transaction.Transaction, messageID uint64) *errs.Error {
	oldMessage, internalErr := m.FindMessageByIDWithTx(ct, tx, messageID)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := m.db.GetTable(MessageTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currMessage := row.(entity.Message)
				if currMessage.ID != messageID {
					rows = append(rows, currMessage)
				}
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err := m.db.GetTable(MessageTableName)
			if err != nil {
				return err
			}

			table.Rows = append(table.Rows, oldMessage)
			return nil
		},
	})
}

func NewMessage(db *dbtest.InMemoryDB, transactionFactory transaction.Factory) Message {
	return Message{db: db, transactionFactory: transactionFactory}
}
