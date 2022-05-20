package sqldb

import (
	"database/sql"

	"github.com/teamyapp/teamy-backend/app/dao"
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type Task struct {
	db *sql.DB
}

var _ dao.Task = (*Task)(nil)

func (t Task) FindAllTasks() ([]entityv2.Task, error) {
	//TODO implement me
	panic("implement me")
}

func (t Task) FindTasksByTeamID(teamID uint64) ([]entityv2.Task, error) {
	//TODO implement me
	panic("implement me")
}

func (t Task) CreateTask(task entityv2.Task) error {
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
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
`,
		task.ID,
		task.Goal,
		task.Context,
		task.CreatorID,
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

func NewTask(sqlDB *sql.DB) Task {
	return Task{db: sqlDB}
}
