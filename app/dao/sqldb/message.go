package sqldb

import (
	"database/sql"
	"log"

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
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	messages := make([]entityv2.Message, 0)
	for rows.Next() {
		message := entityv2.Message{}
		err = rows.Scan(
			&message.ID,
			&message.Body,
			&message.ThreadID,
			&message.AuthorUserID,
			&message.CreatedAt,
			&message.UpdatedAt,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		messages = append(messages, message)
	}

	return messages, err
}

func (m Message) CreateMessage(message entityv2.Message) error {
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
	return err
}

func NewMessage(sqlDB *sql.DB) Message {
	return Message{db: sqlDB}
}
