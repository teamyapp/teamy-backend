package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/teamyapp/teamy-backend/app/dao"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Task struct {
	db *sql.DB
}

var _ dao.Task = (*Task)(nil)

func (t Task) FindTaskByID(taskID uint64) (entity.Task, error) {
	task := entity.Task{}
	err := t.db.QueryRow(`
		SELECT
			id,
			goal,
			context,
			creator_user_id,
			owner_user_id,
			owning_team_id,
			status,
			effort,
			comments_thread_id,
			due_at,
			created_at,
			updated_at
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
			&task.Effort,
			&task.CommentsThreadID,
			&task.DueAt,
			&task.CreatedAt,
			&task.UpdatedAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.Task{}, dao.ErrNotFound(fmt.Sprintf(
			"task not found: id=%v",
			taskID))
	}

	return task, err
}

func (t Task) FindAllTasks() ([]entity.Task, error) {
	statement := `
	SELECT
		id,
		goal,
		context,
		creator_user_id,
		owner_user_id,
		owning_team_id,
		status,
		effort,
		comments_thread_id,
		due_at,
		created_at,
		updated_at
	FROM task;
`
	rows, err := t.db.Query(statement)
	if err != nil {
		log.Println(err)
		return nil, err
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
			&task.Effort,
			&task.CommentsThreadID,
			&task.DueAt,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		tasks = append(tasks, task)
	}

	return tasks, err
}

func (t Task) FindTasksByTeamID(teamID uint64) ([]entity.Task, error) {
	statement := `
	SELECT
		id,
		goal,
		context,
		creator_user_id,
		owner_user_id,
		owning_team_id,
		status,
		effort,
		comments_thread_id,
		due_at,
		created_at,
		updated_at
	FROM task
	WHERE owning_team_id = $1;
`
	rows, err := t.db.Query(statement, teamID)
	if err != nil {
		log.Println(err)
		return nil, err
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
			&task.Effort,
			&task.CommentsThreadID,
			&task.DueAt,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		tasks = append(tasks, task)
	}

	return tasks, err
}

func (t Task) CreateTask(task entity.Task) error {
	_, err := t.db.Exec(`
		INSERT INTO task
		(
			id,
			goal,
			context,
			creator_user_id,
			owner_user_id,
			owning_team_id,
			status,
			effort,
			comments_thread_id,
			due_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`,
		task.ID,
		task.Goal,
		task.Context,
		task.CreatorUserID,
		task.OwnerUserID,
		task.OwningTeamID,
		task.Status,
		task.Effort,
		task.CommentsThreadID,
		task.DueAt,
		task.CreatedAt,
	)
	return err
}

func (t Task) UpdateTask(task entity.Task) error {
	_, err := t.db.Exec(`
		UPDATE task
		SET
			goal = $1,
			context = $2,
			owner_user_id = $3,
			owning_team_id = $4,
			status = $5,
			effort = $6,
			due_at = $7,
			updated_at = $8
		WHERE id = $9;`,
		task.Goal,
		task.Context,
		task.OwnerUserID,
		task.OwningTeamID,
		task.Status,
		task.Effort,
		task.DueAt,
		task.UpdatedAt,
		task.ID,
	)
	return err
}

func (t Task) DeleteTask(taskID uint64) error {
	_, err := t.db.Exec(`
		DELETE FROM task
		WHERE id = $1;
		`,
		taskID)
	return err
}

func NewTask(sqlDB *sql.DB) Task {
	return Task{db: sqlDB}
}
