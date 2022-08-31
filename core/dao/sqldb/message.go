package sqldb

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Message struct {
	dataCollector obs.DataCollector
	db            *sql.DB
}

var _ dao.Message = (*Message)(nil)

func (m Message) FindMessageByID(messageID uint64) (entity.Message, error) {
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
		return entity.Message{}, dao.ErrNotFound(fmt.Sprintf(
			"message not found: id=%v",
			messageID))
	}

	if err != nil {
		m.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return message, err
}

func (m Message) FindMessagesByThreadID(threadID uint64) ([]entity.Message, error) {
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
		m.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
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
			m.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		messages = append(messages, message)
	}

	return messages, nil
}

func (m Message) CreateMessage(message entity.Message) error {
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
		m.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (m Message) UpdateMessage(message entity.Message) error {
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
		m.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (m Message) DeleteMessage(messageID uint64) error {
	_, err := m.db.Exec(`
		DELETE FROM message
		WHERE id = $1;
		`,
		messageID)

	if err != nil {
		m.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func NewMessage(dataCollector obs.DataCollector, sqlDB *sql.DB) Message {
	return Message{dataCollector: dataCollector, db: sqlDB}
}
