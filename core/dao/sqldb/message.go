package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Message struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
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
		internalErr := &errs.Error{
			Code: errs.NotFound,
			Message: fmt.Sprintf(
				"message not found: messageID=%v", messageID),
		}
		m.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.Message{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		m.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.Message{}, internalErr
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		m.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
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
			continue
		}

		messages = append(messages, message)
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		m.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	return messages, nil
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		m.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		m.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		m.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func NewMessage(dataCollector telemetry.DataCollector, sqlDB *sql.DB) Message {
	return Message{dataCollector: dataCollector, db: sqlDB}
}
