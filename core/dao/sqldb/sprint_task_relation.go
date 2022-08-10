package sqldb

import (
	"context"
	"database/sql"
	"log"

	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintTaskRelation struct {
	db *sql.DB
}

var _ dao.SprintTaskRelation = (*SprintTaskRelation)(nil)

func (s SprintTaskRelation) FindTaskIDsBySprintID(sprintID uint64) ([]uint64, error) {
	rows, err := s.db.Query(
		`
	SELECT
		task_id
	FROM sprint_task_relation
	WHERE sprint_id = $1;
`,
		sprintID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()
	taskIDs := make([]uint64, 0)
	for rows.Next() {
		var taskID uint64
		err = rows.Scan(
			&taskID,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		taskIDs = append(taskIDs, taskID)
	}

	return taskIDs, err
}

func (s SprintTaskRelation) FindSprintIDsByTaskID(taskID uint64) ([]uint64, error) {
	rows, err := s.db.Query(
		`
	SELECT
		sprint_id
	FROM sprint_task_relation
	WHERE task_id = $1;
`,
		taskID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()
	sprintIDs := make([]uint64, 0)
	for rows.Next() {
		var sprintID uint64
		err = rows.Scan(
			&sprintID,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		sprintIDs = append(sprintIDs, sprintID)
	}

	return sprintIDs, err
}

func (s SprintTaskRelation) CreateSprintTaskRelation(relation entity.SprintTaskRelation) error {
	_, err := s.db.Exec(`
		INSERT INTO sprint_task_relation
		(
			sprint_id,
			task_id,
			created_at
		)
		VALUES ($1, $2, $3);`,
		relation.SprintID,
		relation.TaskID,
		relation.CreatedAt,
	)
	return err
}

func (s SprintTaskRelation) DeleteSprintTaskRelation(sprintID uint64, taskID uint64) error {
	_, err := s.db.Exec(`
		DELETE FROM sprint_task_relation
		WHERE sprint_id = $1 AND task_id = $2;
		`,
		sprintID,
		taskID)
	return err
}

func (s SprintTaskRelation) MoveTaskToSprint(ct context.Context, fromSprintID uint64, relation entity.SprintTaskRelation, taskID uint64) error {
	// Get a Tx for making transaction requests.
	tx, err := s.db.BeginTx(ct, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	// Defer a rollback in case anything fails.
	defer tx.Rollback()
	err = s.DeleteSprintTaskRelation(fromSprintID, taskID)

	if err != nil {
		log.Println(err)
		return err
	}

	s.CreateSprintTaskRelation(relation)

	if err != nil {
		log.Println(err)
		return err
	}

	if err = tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func NewSprintTaskRelation(db *sql.DB) SprintTaskRelation {
	return SprintTaskRelation{db: db}
}
