package sqldb

import (
	"database/sql"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskAwaitForRelation struct {
	dataCollector obs.DataCollector
	db            *sql.DB
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
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	defer rows.Close()
	waitingTaskIDs := make([]uint64, 0)
	for rows.Next() {
		var waitingTaskID uint64
		err = rows.Scan(&waitingTaskID)
		if err != nil {
			t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		waitingTaskIDs = append(waitingTaskIDs, waitingTaskID)
	}

	return waitingTaskIDs, nil
}

func (t TaskAwaitForRelation) FindAwaitForTaskIDs(waitingTaskID uint64) ([]uint64, error) {
	rows, err := t.db.Query(`
	SELECT
		await_for_task_id
	FROM task_await_for_relation
	WHERE awaiting_task_id = $1;
`, waitingTaskID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	defer rows.Close()
	waitForTaskIDs := make([]uint64, 0)
	for rows.Next() {
		var waitForTaskID uint64
		err = rows.Scan(&waitForTaskID)
		if err != nil {
			t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		waitForTaskIDs = append(waitForTaskIDs, waitForTaskID)
	}

	return waitForTaskIDs, nil
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
		relation.AwaitingTaskID,
		relation.AwaitForTaskID,
		relation.CreatedAt,
	)

	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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

	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func NewTaskAwaitForRelation(dataCollector obs.DataCollector, db *sql.DB) TaskAwaitForRelation {
	return TaskAwaitForRelation{dataCollector: dataCollector, db: db}
}
