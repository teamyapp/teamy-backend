package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Message struct {
	db *sql.DB
}

var _ dao.Message = (*Message)(nil)

func (m Message) FindMessageByID(ct context.Context, messageID uint64) (entity.Message, *errs.Error) {
	message := entity.Message{}
	err := m.db.QueryRow(`
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
	if errors.Is(err, sql.ErrNoRows) {
		return entity.Message{}, errs.NewError(errs.NotFound,
			fmt.Sprintf(
				"message not found: messageID=%v", messageID))
	}

	if err != nil {
		return entity.Message{}, errs.NewError(errs.Unknown, err.Error())
	}

	return message, nil
}

func (m Message) FindMessagesByThreadID(ct context.Context, threadID uint64) ([]entity.Message, *errs.Error) {
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
	rows, err := m.db.Query(statement, threadID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	var internalErr *errs.Error
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

	return messages, internalErr
}

func (m Message) CreateMessage(ct context.Context, message entity.Message) *errs.Error {
	_, err := m.db.Exec(`
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

func (m Message) UpdateMessage(ct context.Context, message entity.Message) *errs.Error {
	_, err := m.db.Exec(`
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

func (m Message) DeleteMessage(ct context.Context, messageID uint64) *errs.Error {
	_, err := m.db.Exec(`
		DELETE FROM message
		WHERE id = $1;
		`,
		messageID)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewMessage(sqlDB *sql.DB) Message {
	return Message{db: sqlDB}
}
