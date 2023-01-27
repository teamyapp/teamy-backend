package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskLink struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.TaskLink = (*TaskLink)(nil)

func (t TaskLink) CreateTaskLink(ct context.Context, taskLinkEntity entity.TaskLink) error {
	_, err := t.db.Exec(`
		INSERT INTO task_link
		(
			id,
			task_id,
			title,
			url,
			icon_url,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6);`,
		taskLinkEntity.ID,
		taskLinkEntity.TaskID,
		taskLinkEntity.Title,
		taskLinkEntity.URL,
		taskLinkEntity.IconURL,
		taskLinkEntity.CreatedAt,
	)

	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return err
}

func (t TaskLink) FindLinksByTaskID(ct context.Context, taskID uint64) ([]entity.TaskLink, error) {
	rows, err := t.db.Query(`
	SELECT
		id,
		task_id,
		title,
		url,
		icon_url,
		created_at,
		updated_at,
	FROM task_link
	WHERE task_id = $1;
`, taskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	defer rows.Close()
	taskLinks := make([]entity.TaskLink, 0)
	for rows.Next() {
		taskLink := entity.TaskLink{}
		err = rows.Scan(
			&taskLink.ID,
			&taskLink.TaskID,
			&taskLink.Title,
			&taskLink.URL,
			&taskLink.IconURL,
			&taskLink.CreatedAt,
			&taskLink.UpdatedAt,
		)
		if err != nil {
			t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			continue
		}

		taskLinks = append(taskLinks, taskLink)
	}

	return taskLinks, nil
}

func NewTaskLink(dataCollector telemetry.DataCollector, sqlDB *sql.DB) TaskLink {
	return TaskLink{dataCollector: dataCollector, db: sqlDB}
}
