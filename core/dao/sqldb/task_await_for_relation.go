package sqldb

import (
	"database/sql"
	"log"

	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskAwaitForRelation struct {
	db *sql.DB
}

var _ dao.TaskAwaitForRelation = (*TaskAwaitForRelation)(nil)

func (t TaskAwaitForRelation) FindAwaitingTaskIDs(waitForTaskID uint64) ([]uint64, error) {
	rows, err := t.db.Query(`
	SELECT
		awaiting_task_id
	FROM task_await_for_relation
	WHERE await_for_task_id = $1;
`, waitForTaskID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()
	waitingTaskIDs := make([]uint64, 0)
	for rows.Next() {
		var waitingTaskID uint64
		err = rows.Scan(&waitingTaskID)
		if err != nil {
			log.Println(err)
			continue
		}

		waitingTaskIDs = append(waitingTaskIDs, waitingTaskID)
	}

	return waitingTaskIDs, err
}

func (t TaskAwaitForRelation) FindAwaitForTaskIDs(waitingTaskID uint64) ([]uint64, error) {
	rows, err := t.db.Query(`
	SELECT
		await_for_task_id
	FROM task_await_for_relation
	WHERE awaiting_task_id = $1;
`, waitingTaskID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()
	waitForTaskIDs := make([]uint64, 0)
	for rows.Next() {
		var waitForTaskID uint64
		err = rows.Scan(&waitForTaskID)
		if err != nil {
			log.Println(err)
			continue
		}

		waitForTaskIDs = append(waitForTaskIDs, waitForTaskID)
	}

	return waitForTaskIDs, err
}

func (t TaskAwaitForRelation) CreateRelation(relation entity.TaskAwaitForRelation) error {
	_, err := t.db.Exec(`
	INSERT INTO task_await_for_relation
	(
	    awaiting_task_id,
	    await_for_task_id,
	 	created_at
	)
	VALUES ($1, $2, $3);
`,
		relation.AWaitingTaskID,
		relation.AWaitForTaskID,
		relation.CreatedAt,
	)

	if err != nil {
		log.Println(err)
	}

	return err
}

func (t TaskAwaitForRelation) DeleteRelation(waitingTaskID uint64, awaitForTaskID uint64) error {
	_, err := t.db.Exec(`
		DELETE FROM task_await_for_relation
		WHERE awaiting_task_id = $1 AND await_for_task_id = $2;
		`,
		waitingTaskID,
		awaitForTaskID)
	return err
}

func NewTaskAwaitForRelation(db *sql.DB) TaskAwaitForRelation {
	return TaskAwaitForRelation{
		db: db,
	}
}
