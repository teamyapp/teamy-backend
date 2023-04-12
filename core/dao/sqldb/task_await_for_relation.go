package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskAwaitForRelation struct {
	db *sql.DB
}

var _ dao.TaskAwaitForRelation = (*TaskAwaitForRelation)(nil)

func (t TaskAwaitForRelation) FindAwaitingTaskIDs(ct context.Context, waitForTaskID uint64) ([]uint64, *errs.Error) {
	rows, err := t.db.Query(`
	SELECT
		awaiting_task_id
	FROM task_await_for_relation
	WHERE await_for_task_id = $1;
`, waitForTaskID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	waitingTaskIDs := make([]uint64, 0)
	for rows.Next() {
		var waitingTaskID uint64
		err = rows.Scan(&waitingTaskID)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		waitingTaskIDs = append(waitingTaskIDs, waitingTaskID)
	}

	return waitingTaskIDs, nil
}

func (t TaskAwaitForRelation) FindAwaitForTaskIDs(ct context.Context, waitingTaskID uint64) ([]uint64, *errs.Error) {
	rows, err := t.db.Query(`
	SELECT
		await_for_task_id
	FROM task_await_for_relation
	WHERE awaiting_task_id = $1;
`, waitingTaskID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	waitForTaskIDs := make([]uint64, 0)
	for rows.Next() {
		var waitForTaskID uint64
		err = rows.Scan(&waitForTaskID)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		waitForTaskIDs = append(waitForTaskIDs, waitForTaskID)
	}

	return waitForTaskIDs, nil
}

func (t TaskAwaitForRelation) CreateRelation(ct context.Context, relation entity.TaskAwaitForRelation) *errs.Error {
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
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t TaskAwaitForRelation) DeleteRelation(ct context.Context, waitingTaskID uint64, awaitForTaskID uint64) *errs.Error {
	_, err := t.db.Exec(`
		DELETE FROM task_await_for_relation
		WHERE awaiting_task_id = $1 AND await_for_task_id = $2;
		`,
		waitingTaskID,
		awaitForTaskID)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewTaskAwaitForRelation(db *sql.DB) TaskAwaitForRelation {
	return TaskAwaitForRelation{db: db}
}
