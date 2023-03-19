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

type TaskLink struct {
	dataCollector telemetry.DataCollector
}

var _ daov2.TaskLink = (*TaskLink)(nil)

func (t TaskLink) CreateTaskLink(ct context.Context, tx *transaction.Transaction, taskLinkEntity entity.TaskLink) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		INSERT INTO task_link
		(
			id,
			task_id,
			title,
			url,
			icon_url,
			icon_hover_url,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7);`,
		taskLinkEntity.ID,
		taskLinkEntity.TaskID,
		taskLinkEntity.Title,
		taskLinkEntity.URL,
		taskLinkEntity.IconURL,
		taskLinkEntity.IconHoverURL,
		taskLinkEntity.CreatedAt,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t TaskLink) DeleteTaskLink(ct context.Context, tx *transaction.Transaction, taskLinkID uint64) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		DELETE FROM task_link
		WHERE id = $1;
		`,
		taskLinkID)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t TaskLink) FindTaskLinkByID(ct context.Context, tx *transaction.Transaction, taskLinkID uint64) (entity.TaskLink, *errs.Error) {
	taskLink := entity.TaskLink{}
	err := tx.SQLTx().QueryRow(`
		SELECT
			id,
			task_id,
			title,
			url,
			icon_url,
			icon_hover_url,
			created_at,
			updated_at
		FROM task_link
		WHERE id = $1;`,
		taskLinkID).
		Scan(
			&taskLink.ID,
			&taskLink.TaskID,
			&taskLink.Title,
			&taskLink.URL,
			&taskLink.IconURL,
			&taskLink.IconHoverURL,
			&taskLink.CreatedAt,
			&taskLink.UpdatedAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.TaskLink{}, errs.NewError(errs.NotFound, fmt.Sprintf("taskLink not found: id=%v", taskLinkID))
	}

	if err != nil {
		return taskLink, errs.NewError(errs.Unknown, err.Error())
	}

	return taskLink, nil
}

func (t TaskLink) FindLinksByTaskID(ct context.Context, tx *transaction.Transaction, taskID uint64) ([]entity.TaskLink, *errs.Error) {
	query := fmt.Sprintf(`
	SELECT
		id,
		task_id,
		title,
		url,
		icon_url,
		icon_hover_url,
		created_at,
		updated_at
	FROM task_link
	WHERE task_id = %d;
`, taskID)

	rows, err := tx.SQLTx().Query(query)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	var internalErr *errs.Error
	taskLinks := make([]entity.TaskLink, 0)
	for rows.Next() {
		taskLink := entity.TaskLink{}
		err = rows.Scan(
			&taskLink.ID,
			&taskLink.TaskID,
			&taskLink.Title,
			&taskLink.URL,
			&taskLink.IconURL,
			&taskLink.IconHoverURL,
			&taskLink.CreatedAt,
			&taskLink.UpdatedAt,
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

		taskLinks = append(taskLinks, taskLink)
	}

	return taskLinks, nil
}

func NewTaskLink(dataCollector telemetry.DataCollector, sqlDB *sql.DB) TaskLink {
	return TaskLink{dataCollector: dataCollector}
}
