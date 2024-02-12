package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Message struct {
	transactionFactory transaction.Factory
}

var _ dao.Message = (*Message)(nil)

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
	message := entity.Message{}
	err := tx.SQLTx().QueryRow(`
		SELECT
			id,
			body,
			thread_id,
			author_user_id,
			created_at,
			updated_at
		FROM message
		WHERE id = $1;`,
		messageID).
		Scan(
			&message.ID,
			&message.Body,
			&message.ThreadID,
			&message.AuthorUserID,
			&message.CreatedAt,
			&message.UpdatedAt,
		)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Message{}, errs.NewError(errs.NotFound, fmt.Sprintf(
				"message not found: messageID=%v", messageID))
		}

		return entity.Message{}, errs.NewError(errs.Unknown, err.Error())
	}

	return message, nil
}

func (m Message) FindMessagesByThreadIDWithTx(ct context.Context, tx *transaction.Transaction, threadID uint64) ([]entity.Message, *errs.Error) {
	statement := `
	SELECT
		id,
		body,
		thread_id,
		author_user_id,
		created_at,
		updated_at
	FROM message
	WHERE thread_id = $1;
`
	rows, err := tx.SQLTx().Query(statement, threadID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	messages := make([]entity.Message, 0)
	for rows.Next() {
		message := entity.Message{}
		err = rows.Scan(
			&message.ID,
			&message.Body,
			&message.ThreadID,
			&message.AuthorUserID,
			&message.CreatedAt,
			&message.UpdatedAt,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		messages = append(messages, message)
	}

	return messages, nil
}

func (m Message) CreateMessage(ct context.Context, tx *transaction.Transaction, message entity.Message) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		INSERT INTO message
		(
			id,
			body,
			thread_id,
			author_user_id,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5);`,
		message.ID,
		message.Body,
		message.ThreadID,
		message.AuthorUserID,
		message.CreatedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (m Message) UpdateMessage(ct context.Context, tx *transaction.Transaction, message entity.Message) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		UPDATE message
		SET
			body = $1,
			updated_at = $2
		WHERE id = $3;`,
		message.Body,
		message.UpdatedAt,
		message.ID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (m Message) DeleteMessage(ct context.Context, tx *transaction.Transaction, messageID uint64) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		DELETE FROM message
		WHERE id = $1;
		`,
		messageID)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewMessage(transactionFactory transaction.Factory) Message {
	return Message{transactionFactory: transactionFactory}
}
