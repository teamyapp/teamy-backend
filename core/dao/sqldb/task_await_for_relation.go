package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const taskAwaitForRelationDaoName = "TaskAwaitForRelation"

type TaskAwaitForRelation struct {
	metrics dao.Metrics
}

var _ dao.TaskAwaitForRelation = (*TaskAwaitForRelation)(nil)

func (t TaskAwaitForRelation) FindAwaitingTaskIDsWithTx(ct context.Context, tx *transaction.Transaction, waitForTaskID uint64) ([]uint64, *errs.Error) {
	t.metrics.ReportDaoOperation(taskAwaitForRelationDaoName, "FindAwaitingTaskIDsWithTx")
	rows, err := tx.SQLTx().Query(`
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

func (t TaskAwaitForRelation) FindAwaitForTaskIDsWithTx(ct context.Context, tx *transaction.Transaction, waitingTaskID uint64) ([]uint64, *errs.Error) {
	t.metrics.ReportDaoOperation(taskAwaitForRelationDaoName, "FindAwaitForTaskIDsWithTx")
	rows, err := tx.SQLTx().Query(`
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

func (t TaskAwaitForRelation) CreateRelation(ct context.Context, tx *transaction.Transaction, relation entity.TaskAwaitForRelation) *errs.Error {
	t.metrics.ReportDaoOperation(taskAwaitForRelationDaoName, "CreateRelation")
	_, err := tx.SQLTx().Exec(`
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

func (t TaskAwaitForRelation) DeleteRelation(ct context.Context, tx *transaction.Transaction, waitingTaskID uint64, awaitForTaskID uint64) *errs.Error {
	t.metrics.ReportDaoOperation(taskAwaitForRelationDaoName, "DeleteRelation")
	_, err := tx.SQLTx().Exec(`
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

func NewTaskAwaitForRelation(metrics dao.Metrics) TaskAwaitForRelation {
	return TaskAwaitForRelation{
		metrics: metrics,
	}
}
