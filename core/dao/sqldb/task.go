package sqldb

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Task struct {
	dataCollector obs.DataCollector
	db            *sql.DB
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
			is_planned,
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
			&task.IsPlanned,
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

	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return task, err
}

func (t Task) FindTasksByIDs(taskIDs []uint64) ([]entity.Task, error) {
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
		updated_at
	FROM task
	WHERE id IN (%s);`, idsString)
	rows, err := t.db.Query(query)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
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
			)
		if err != nil {
			t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (t Task) FindTaskByCommentsThreadID(commentThreadID uint64) (entity.Task, error) {
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
			is_planned,
			effort,
			comments_thread_id,
			due_at,
			created_at,
			updated_at
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
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.Task{}, dao.ErrNotFound(fmt.Sprintf(
			"task not found: commentsThreadID=%v",
			commentThreadID))
	}

	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return task, err
}

func (t Task) FindAllTasks() ([]entity.Task, error) {
	rows, err := t.db.Query(`
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
		updated_at
	FROM task;
`)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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
			&task.IsPlanned,
			&task.Effort,
			&task.CommentsThreadID,
			&task.DueAt,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (t Task) FindTasksByTeamID(teamID uint64) ([]entity.Task, error) {
	rows, err := t.db.Query(
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
		updated_at
	FROM task
	WHERE owning_team_id = $1;
`,
		teamID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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
			&task.IsPlanned,
			&task.Effort,
			&task.CommentsThreadID,
			&task.DueAt,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
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
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

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
			is_planned = $6,
			effort = $7,
			due_at = $8,
			updated_at = $9
		WHERE id = $10;`,
		task.Goal,
		task.Context,
		task.OwnerUserID,
		task.OwningTeamID,
		task.Status,
		task.IsPlanned,
		task.Effort,
		task.DueAt,
		task.UpdatedAt,
		task.ID,
	)

	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (t Task) DeleteTask(taskID uint64) error {
	_, err := t.db.Exec(`
		DELETE FROM task
		WHERE id = $1;
		`,
		taskID)

	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func NewTask(dataCollector obs.DataCollector, sqlDB *sql.DB) Task {
	return Task{dataCollector: dataCollector, db: sqlDB}
}
