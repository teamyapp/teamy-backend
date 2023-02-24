package implementation

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskAwaitForRelation struct {
	dataCollector telemetry.DataCollector
}

var _ daov2.TaskAwaitForRelation = (*TaskAwaitForRelation)(nil)

func (t TaskAwaitForRelation) FindAwaitingTaskIDs(ct context.Context, tx *sql.Tx, waitForTaskID uint64) ([]uint64, *errs.Error) {
	rows, err := tx.Query(`
	SELECT
		awaiting_task_id
	FROM task_await_for_relation
	WHERE await_for_task_id = $1;
`, waitForTaskID)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
	waitingTaskIDs := make([]uint64, 0)
	for rows.Next() {
		var waitingTaskID uint64
		err = rows.Scan(&waitingTaskID)
		if err != nil {
			newInternalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}

			if internalErr == nil {
				internalErr = newInternalErr
			}

			t.dataCollector.Logger.ErrorWithContext(ct, newInternalErr)
			continue
		}

		waitingTaskIDs = append(waitingTaskIDs, waitingTaskID)
	}

	return waitingTaskIDs, nil
}

func (t TaskAwaitForRelation) FindAwaitForTaskIDs(ct context.Context, tx *sql.Tx, waitingTaskID uint64) ([]uint64, *errs.Error) {
	rows, err := tx.Query(`
	SELECT
		await_for_task_id
	FROM task_await_for_relation
	WHERE awaiting_task_id = $1;
`, waitingTaskID)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
	waitForTaskIDs := make([]uint64, 0)
	for rows.Next() {
		var waitForTaskID uint64
		err = rows.Scan(&waitForTaskID)
		if err != nil {
			newInternalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}

			if internalErr == nil {
				internalErr = newInternalErr
			}

			t.dataCollector.Logger.ErrorWithContext(ct, newInternalErr)
			continue
		}

		waitForTaskIDs = append(waitForTaskIDs, waitForTaskID)
	}

	return waitForTaskIDs, nil
}

func (t TaskAwaitForRelation) CreateRelation(ct context.Context, tx *sql.Tx, relation entity.TaskAwaitForRelation) *errs.Error {
	_, err := tx.Exec(`
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (t TaskAwaitForRelation) DeleteRelation(ct context.Context, tx *sql.Tx, waitingTaskID uint64, awaitForTaskID uint64) *errs.Error {
	_, err := tx.Exec(`
		DELETE FROM task_await_for_relation
		WHERE awaiting_task_id = $1 AND await_for_task_id = $2;
		`,
		waitingTaskID,
		awaitForTaskID)

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func NewTaskAwaitForRelation(dataCollector telemetry.DataCollector) TaskAwaitForRelation {
	return TaskAwaitForRelation{dataCollector: dataCollector}
}
