package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintTaskRelation struct {
	db *sql.DB
}

var _ dao.SprintTaskRelation = (*SprintTaskRelation)(nil)

func (s SprintTaskRelation) FindTaskIDsBySprintID(ct context.Context, sprintID uint64) ([]uint64, *errs.Error) {
	rows, err := s.db.Query(
		`
	SELECT
		task_id
	FROM sprint_task_relation
	WHERE sprint_id = $1;
`,
		sprintID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
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
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		taskIDs = append(taskIDs, taskID)
	}

	return taskIDs, internalErr
}

func (s SprintTaskRelation) FindSprintIDsByTaskID(ct context.Context, taskID uint64) ([]uint64, *errs.Error) {
	rows, err := s.db.Query(
		`
	SELECT
		sprint_id
	FROM sprint_task_relation
	WHERE task_id = $1;
`,
		taskID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
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
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		sprintIDs = append(sprintIDs, sprintID)
	}

	return sprintIDs, internalErr
}

func (s SprintTaskRelation) CreateSprintTaskRelation(ct context.Context, relation entity.SprintTaskRelation) *errs.Error {
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
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (s SprintTaskRelation) DeleteSprintTaskRelation(ct context.Context, sprintID uint64, taskID uint64) *errs.Error {
	_, err := s.db.Exec(`
		DELETE FROM sprint_task_relation
		WHERE sprint_id = $1 AND task_id = $2;
		`,
		sprintID,
		taskID)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewSprintTaskRelation(db *sql.DB) SprintTaskRelation {
	return SprintTaskRelation{db: db}
}
