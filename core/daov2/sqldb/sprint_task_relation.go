package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintTaskRelation struct {
	logger telemetry.Logger
}

var _ daov2.SprintTaskRelation = (*SprintTaskRelation)(nil)

func (s SprintTaskRelation) FindTaskIDsBySprintIDWithTx(ct context.Context, tx *transaction.Transaction, sprintID uint64) ([]uint64, *errs.Error) {
	rows, err := tx.SQLTx().Query(
		`
	SELECT
		task_id
	FROM sprint_task_relation
	WHERE sprint_id = $1;
`,
		sprintID)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		s.logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
	taskIDs := make([]uint64, 0)
	for rows.Next() {
		var taskID uint64
		err = rows.Scan(
			&taskID,
		)
		if err != nil {
			newInternalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}

			if internalErr == nil {
				internalErr = newInternalErr
			}

			s.logger.ErrorWithContext(ct, newInternalErr)
			continue
		}

		taskIDs = append(taskIDs, taskID)
	}

	return taskIDs, internalErr
}

func (s SprintTaskRelation) FindSprintIDsByTaskIDWithTx(ct context.Context, tx *transaction.Transaction, taskID uint64) ([]uint64, *errs.Error) {
	rows, err := tx.SQLTx().Query(
		`
	SELECT
		sprint_id
	FROM sprint_task_relation
	WHERE task_id = $1;
`,
		taskID)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		s.logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
	sprintIDs := make([]uint64, 0)
	for rows.Next() {
		var sprintID uint64
		err = rows.Scan(
			&sprintID,
		)
		if err != nil {
			newInternalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}

			if internalErr == nil {
				internalErr = newInternalErr
			}

			s.logger.ErrorWithContext(ct, newInternalErr)
			continue
		}

		sprintIDs = append(sprintIDs, sprintID)
	}

	return sprintIDs, internalErr
}

func (s SprintTaskRelation) CreateSprintTaskRelation(ct context.Context, tx *transaction.Transaction, relation entity.SprintTaskRelation) *errs.Error {
	_, err := tx.SQLTx().Exec(`
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

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		s.logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (s SprintTaskRelation) DeleteSprintTaskRelation(ct context.Context, tx *transaction.Transaction, sprintID uint64, taskID uint64) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		DELETE FROM sprint_task_relation
		WHERE sprint_id = $1 AND task_id = $2;
		`,
		sprintID,
		taskID)

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		s.logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func NewSprintTaskRelation(logger telemetry.Logger) SprintTaskRelation {
	return SprintTaskRelation{logger: logger}
}
