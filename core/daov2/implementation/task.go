package implementation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Task struct {
	dataCollector telemetry.DataCollector
}

var _ daov2.Task = (*Task)(nil)

func (t Task) FindTaskByID(ct context.Context, tx *sql.Tx, taskID uint64) (entity.Task, *errs.Error) {
	task := entity.Task{}
	err := tx.QueryRow(`
		SELECT
			id,
			goal,
			context,
			creator_user_id,
			owner_user_id,
			owning_team_id,
			status,
			is_planned,
			effort,
			comments_thread_id,
			due_at,
			created_at,
			updated_at,
			delivered_at
		FROM task
		WHERE id = $1;`,
		taskID).
		Scan(
			&task.ID,
			&task.Goal,
			&task.Context,
			&task.CreatorUserID,
			&task.OwnerUserID,
			&task.OwningTeamID,
			&task.Status,
			&task.IsPlanned,
			&task.Effort,
			&task.CommentsThreadID,
			&task.DueAt,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.DeliveredAt,
		)
	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("task not found: taskID=%v", taskID),
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Task{}, internalErr
	}

	return task, nil
}

func (t Task) FindTasksByIDs(ct context.Context, tx *sql.Tx, taskIDs []uint64) ([]entity.Task, *errs.Error) {
	if len(taskIDs) == 0 {
		return []entity.Task{}, nil
	}

	idsString := toIDsString(taskIDs)
	query := fmt.Sprintf(`
	SELECT
		id,
		goal,
		context,
		creator_user_id,
		owner_user_id,
		owning_team_id,
		status,
		is_planned,
		effort,
		comments_thread_id,
		due_at,
		created_at,
		updated_at,
		delivered_at
	FROM task
	WHERE id IN (%v);`, idsString)
	rows, err := tx.Query(query)
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
	var tasks []entity.Task
	for rows.Next() {
		var task entity.Task
		err = rows.
			Scan(
				&task.ID,
				&task.Goal,
				&task.Context,
				&task.CreatorUserID,
				&task.OwnerUserID,
				&task.OwningTeamID,
				&task.Status,
				&task.IsPlanned,
				&task.Effort,
				&task.CommentsThreadID,
				&task.DueAt,
				&task.CreatedAt,
				&task.UpdatedAt,
				&task.DeliveredAt,
			)
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

		tasks = append(tasks, task)
	}

	return tasks, internalErr
}

func (t Task) FindTaskByCommentsThreadID(ct context.Context, tx *sql.Tx, commentThreadID uint64) (entity.Task, *errs.Error) {
	task := entity.Task{}
	err := tx.QueryRow(`
		SELECT
			id,
			goal,
			context,
			creator_user_id,
			owner_user_id,
			owning_team_id,
			status,
			is_planned,
			effort,
			comments_thread_id,
			due_at,
			created_at,
			updated_at,
			delivered_at
		FROM task
		WHERE comments_thread_id = $1;`,
		commentThreadID).
		Scan(
			&task.ID,
			&task.Goal,
			&task.Context,
			&task.CreatorUserID,
			&task.OwnerUserID,
			&task.OwningTeamID,
			&task.Status,
			&task.IsPlanned,
			&task.Effort,
			&task.CommentsThreadID,
			&task.DueAt,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.DeliveredAt,
		)
	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("task not found: commentsThreadID=%v", commentThreadID),
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Task{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Task{}, internalErr
	}

	return task, nil
}

func (t Task) FindAllTasks(ct context.Context, tx *sql.Tx) ([]entity.Task, *errs.Error) {
	rows, err := tx.Query(`
	SELECT
		id,
		goal,
		context,
		creator_user_id,
		owner_user_id,
		owning_team_id,
		status,
		is_planned,
		effort,
		comments_thread_id,
		due_at,
		created_at,
		updated_at,
		delivered_at
	FROM task;
`)
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
	tasks := make([]entity.Task, 0)
	for rows.Next() {
		task := entity.Task{}
		err = rows.Scan(
			&task.ID,
			&task.Goal,
			&task.Context,
			&task.CreatorUserID,
			&task.OwnerUserID,
			&task.OwningTeamID,
			&task.Status,
			&task.IsPlanned,
			&task.Effort,
			&task.CommentsThreadID,
			&task.DueAt,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.DeliveredAt,
		)
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

		tasks = append(tasks, task)
	}

	return tasks, internalErr
}

func (t Task) FindTasksByTeamID(ct context.Context, tx *sql.Tx, teamID uint64) ([]entity.Task, *errs.Error) {
	rows, err := tx.Query(
		`
	SELECT
		id,
		goal,
		context,
		creator_user_id,
		owner_user_id,
		owning_team_id,
		status,
		is_planned,
		effort,
		comments_thread_id,
		due_at,
		created_at,
		updated_at,
		delivered_at
	FROM task
	WHERE owning_team_id = $1;
`,
		teamID)
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
	tasks := make([]entity.Task, 0)
	for rows.Next() {
		task := entity.Task{}
		err = rows.Scan(
			&task.ID,
			&task.Goal,
			&task.Context,
			&task.CreatorUserID,
			&task.OwnerUserID,
			&task.OwningTeamID,
			&task.Status,
			&task.IsPlanned,
			&task.Effort,
			&task.CommentsThreadID,
			&task.DueAt,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.DeliveredAt,
		)
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

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (t Task) CreateTask(ct context.Context, tx *sql.Tx, task entity.Task) *errs.Error {
	_, err := tx.Exec(`
		INSERT INTO task
		(
			id,
			goal,
			context,
			creator_user_id,
			owner_user_id,
			owning_team_id,
			status,
		 	is_planned,
			effort,
			comments_thread_id,
			due_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);`,
		task.ID,
		task.Goal,
		task.Context,
		task.CreatorUserID,
		task.OwnerUserID,
		task.OwningTeamID,
		task.Status,
		task.IsPlanned,
		task.Effort,
		task.CommentsThreadID,
		task.DueAt,
		task.CreatedAt,
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

func (t Task) UpdateTask(ct context.Context, tx *sql.Tx, task entity.Task) *errs.Error {
	_, err := tx.Exec(`
		UPDATE task
		SET
			goal = $1,
			context = $2,
			owner_user_id = $3,
			owning_team_id = $4,
			status = $5,
			is_planned = $6,
			effort = $7,
			due_at = $8,
			updated_at = $9,
			delivered_at = $10
		WHERE id = $11;`,
		task.Goal,
		task.Context,
		task.OwnerUserID,
		task.OwningTeamID,
		task.Status,
		task.IsPlanned,
		task.Effort,
		task.DueAt,
		task.UpdatedAt,
		task.DeliveredAt,
		task.ID,
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

func (t Task) DeleteTask(ct context.Context, tx *sql.Tx, taskID uint64) *errs.Error {
	_, err := tx.Exec(`
		DELETE FROM task
		WHERE id = $1;
		`,
		taskID)
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

func NewTask(dataCollector telemetry.DataCollector) Task {
	return Task{dataCollector: dataCollector}
}
