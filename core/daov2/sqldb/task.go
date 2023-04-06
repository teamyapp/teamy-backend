package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Task struct {
	logger             telemetry.Logger
	transactionFactory transaction.Factory
}

var _ daov2.Task = (*Task)(nil)

func (t Task) FindTaskByID(ct context.Context, taskID uint64) (entity.Task, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := t.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.Task{}, err
	}

	defer tx.Rollback()
	return t.FindTaskByIDWithTx(ct, tx, taskID)
}

func (t Task) FindTasksByTeamID(ct context.Context, teamID uint64) ([]entity.Task, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := t.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return t.FindTasksByTeamIDWithTx(ct, tx, teamID)
}

func (t Task) FindAllTasks(ct context.Context) ([]entity.Task, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := t.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return t.FindAllTasksWithTx(ct, tx)
}

func (t Task) FindTaskByIDWithTx(ct context.Context, tx *transaction.Transaction, taskID uint64) (entity.Task, *errs.Error) {
	task := entity.Task{}
	err := tx.SQLTx().QueryRow(`
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
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Task{}, errs.NewError(errs.NotFound, fmt.Sprintf("task not found: taskID=%v", taskID))
		}

		return entity.Task{}, errs.NewError(errs.Unknown, err.Error())
	}

	return task, nil
}

func (t Task) FindTasksByIDsWithTx(ct context.Context, tx *transaction.Transaction, taskIDs []uint64) ([]entity.Task, *errs.Error) {
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
	rows, err := tx.SQLTx().Query(query)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

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
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (t Task) FindTaskByCommentsThreadIDWithTx(ct context.Context, tx *transaction.Transaction, commentThreadID uint64) (entity.Task, *errs.Error) {
	task := entity.Task{}
	err := tx.SQLTx().QueryRow(`
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
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Task{}, errs.NewError(errs.NotFound, fmt.Sprintf("task not found: commentsThreadID=%v", commentThreadID))
		}

		return entity.Task{}, errs.NewError(errs.Unknown, err.Error())
	}

	return task, nil
}

func (t Task) FindAllTasksWithTx(ct context.Context, tx *transaction.Transaction) ([]entity.Task, *errs.Error) {
	rows, err := tx.SQLTx().Query(`
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
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

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
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (t Task) FindTasksByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.Task, *errs.Error) {
	rows, err := tx.SQLTx().Query(
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
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

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
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (t Task) CreateTask(ct context.Context, tx *transaction.Transaction, task entity.Task) *errs.Error {
	_, err := tx.SQLTx().Exec(`
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
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t Task) UpdateTask(ct context.Context, tx *transaction.Transaction, task entity.Task) *errs.Error {
	_, err := tx.SQLTx().Exec(`
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
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t Task) DeleteTask(ct context.Context, tx *transaction.Transaction, taskID uint64) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		DELETE FROM task
		WHERE id = $1;
		`,
		taskID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewTask(logger telemetry.Logger, transactionFactory transaction.Factory) Task {
	return Task{
		logger:             logger,
		transactionFactory: transactionFactory,
	}
}
