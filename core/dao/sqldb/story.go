package sqldb

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const storyDaoName = "Story"

type Story struct {
	transactionFactory transaction.Factory
	metrics            dao.Metrics
}

var _ dao.Story = (*Story)(nil)

func (s *Story) FindStoriesWithTx(ct context.Context, tx *transaction.Transaction) ([]entity.Story, *errs.Error) {
	s.metrics.ReportDaoOperation(storyDaoName, "FindStoriesWithTx")
	var stories []entity.Story
	rows, err := tx.SQLTx().QueryContext(ct, `
		SELECT
			id,
			name,
			owner_id,
			status,
			priority,
			creator_id,
			created_at,
			updated_at,
			is_planned
		FROM story
	`)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		story := entity.Story{}
		err := rows.Scan(
			&story.ID,
			&story.Name,
			&story.OwnerID,
			&story.Status,
			&story.Priority,
			&story.CreatorID,
			&story.CreatedAt,
			&story.UpdatedAt,
			&story.IsPlanned,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		stories = append(stories, story)
	}

	return stories, nil
}

func (s *Story) FindStoriesByIDsWithTx(ct context.Context, tx *transaction.Transaction, storyIDs []uint64) ([]entity.Story, *errs.Error) {
	s.metrics.ReportDaoOperation(storyDaoName, "FindStoriesByIDsWithTx")
	if len(storyIDs) == 0 {
		return []entity.Story{}, nil
	}

	stories := []entity.Story{}
	idsStr := toIDsString(storyIDs)
	query := fmt.Sprintf(`
		SELECT
			id,
			name,
			owner_id,
			status,
			priority,
			creator_id,
			created_at,
			updated_at,
			is_planned
		FROM story
		WHERE id IN (%s)
	`, idsStr)
	rows, err := tx.SQLTx().QueryContext(ct, query)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		story := entity.Story{}
		err := rows.Scan(
			&story.ID,
			&story.Name,
			&story.OwnerID,
			&story.Status,
			&story.Priority,
			&story.CreatorID,
			&story.CreatedAt,
			&story.UpdatedAt,
			&story.IsPlanned,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		stories = append(stories, story)
	}

	return stories, nil
}

func (s *Story) FindStoryByIDWithTx(ct context.Context, tx *transaction.Transaction, storyID uint64) (entity.Story, *errs.Error) {
	s.metrics.ReportDaoOperation(storyDaoName, "FindStoryByIDWithTx")
	story := entity.Story{}
	err := tx.SQLTx().QueryRowContext(ct, `
		SELECT
			id,
			name,
			owner_id,
			status,
			priority,
			creator_id,
			created_at,
			updated_at,
			is_planned
			FROM story
		WHERE id = $1
	`, storyID).Scan(
		&story.ID,
		&story.Name,
		&story.OwnerID,
		&story.Status,
		&story.Priority,
		&story.CreatorID,
		&story.CreatedAt,
		&story.UpdatedAt,
		&story.IsPlanned,
	)

	if err != nil {
		return entity.Story{}, errs.NewError(errs.Unknown, err.Error())
	}

	return story, nil
}

func (s *Story) CreateStory(ct context.Context, tx *transaction.Transaction, story entity.Story) *errs.Error {
	s.metrics.ReportDaoOperation(storyDaoName, "CreateStory")
	_, err := tx.SQLTx().ExecContext(ct, `
		INSERT INTO story (
			id,
			name,
			owner_id,
			status,
			priority,
			creator_id,
			created_at,
			updated_at,
			is_planned
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`,
		story.ID,
		story.Name,
		story.OwnerID,
		story.Status,
		story.Priority,
		story.CreatorID,
		story.CreatedAt,
		story.UpdatedAt,
		story.IsPlanned,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (s *Story) UpdateStory(ct context.Context, tx *transaction.Transaction, story entity.Story) *errs.Error {
	s.metrics.ReportDaoOperation(storyDaoName, "UpdateStory")
	_, err := tx.SQLTx().ExecContext(ct, `
		UPDATE story
		SET
			name = $1,
			owner_id = $2,
			status = $3,
			priority = $4,
			creator_id = $5,
			created_at = $6,
			updated_at = $7,
			is_planned = $8
		WHERE id = $9;
	`,
		story.Name,
		story.OwnerID,
		story.Status,
		story.Priority,
		story.CreatorID,
		story.CreatedAt,
		story.UpdatedAt,
		story.IsPlanned,
		story.ID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (s *Story) DeleteStory(ct context.Context, tx *transaction.Transaction, storyID uint64) *errs.Error {
	s.metrics.ReportDaoOperation(storyDaoName, "DeleteStory")
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM story
		WHERE id = $1;
	`, storyID)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewStory(
	metrics dao.Metrics,
	transactionFactory transaction.Factory,
) *Story {
	return &Story{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
