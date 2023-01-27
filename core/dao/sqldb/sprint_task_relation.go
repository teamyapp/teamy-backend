package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintTaskRelation struct {
	dataCollector obs.DataCollector
	db            *sql.DB
}

var _ dao.SprintTaskRelation = (*SprintTaskRelation)(nil)

func (s SprintTaskRelation) FindTaskIDsBySprintID(ct context.Context, sprintID uint64) ([]uint64, error) {
	rows, err := s.db.Query(
		`
	SELECT
		task_id
	FROM sprint_task_relation
	WHERE sprint_id = $1;
`,
		sprintID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	defer rows.Close()
	taskIDs := make([]uint64, 0)
	for rows.Next() {
		var taskID uint64
		err = rows.Scan(
			&taskID,
		)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		taskIDs = append(taskIDs, taskID)
	}

	return taskIDs, nil
}

func (s SprintTaskRelation) FindSprintIDsByTaskID(ct context.Context, taskID uint64) ([]uint64, error) {
	rows, err := s.db.Query(
		`
	SELECT
		sprint_id
	FROM sprint_task_relation
	WHERE task_id = $1;
`,
		taskID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	defer rows.Close()
	sprintIDs := make([]uint64, 0)
	for rows.Next() {
		var sprintID uint64
		err = rows.Scan(
			&sprintID,
		)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		sprintIDs = append(sprintIDs, sprintID)
	}

	return sprintIDs, nil
}

func (s SprintTaskRelation) CreateSprintTaskRelation(ct context.Context, relation entity.SprintTaskRelation) error {
	_, err := s.db.Exec(`
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
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (s SprintTaskRelation) DeleteSprintTaskRelation(ct context.Context, sprintID uint64, taskID uint64) error {
	_, err := s.db.Exec(`
		DELETE FROM sprint_task_relation
		WHERE sprint_id = $1 AND task_id = $2;
		`,
		sprintID,
		taskID)

	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func NewSprintTaskRelation(dataCollector obs.DataCollector, db *sql.DB) SprintTaskRelation {
	return SprintTaskRelation{dataCollector: dataCollector, db: db}
}
