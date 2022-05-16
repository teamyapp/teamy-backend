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

func NewTask(sqlDB *sql.DB) Task {
	return Task{db: sqlDB}
}
