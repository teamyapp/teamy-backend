package sqldb

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Story struct {
	transactionFactory transaction.Factory
}

var _ dao.Story = (*Story)(nil)

func (s *Story) FindStoriesByIDsWithTx(ct context.Context, tx *transaction.Transaction, storyIDs []uint64) ([]entity.Story, *errs.Error) {
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
		updated_at
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
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}
		
		stories = append(stories, story)
	}

	return stories, nil
}

func (s *Story) FindStoryByIDWithTx(ct context.Context, tx *transaction.Transaction, storyID uint64) (entity.Story, *errs.Error) {
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
		updated_at
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
	)

	if err != nil {
		return entity.Story{}, errs.NewError(errs.Unknown, err.Error())
	}

	return story, nil
}

func (s *Story) CreateStory(ct context.Context, tx *transaction.Transaction, story entity.Story) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		INSERT INTO story (
			id,
			name,
			owner_id,
			status,
			priority,
			creator_id,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`,
		story.ID,
		story.Name,
		story.OwnerID,
		story.Status,
		story.Priority,
		story.CreatorID,
		story.CreatedAt,
		story.UpdatedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (s *Story) UpdateStory(ct context.Context, tx *transaction.Transaction, story entity.Story) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		UPDATE story
		SET
		name = $1,
		owner_id = $2,
		status = $3,
		priority = $4,
		creator_id = $5,
		created_at = $6,
		updated_at = $7
		WHERE id = $8;
	`,
		story.Name,
		story.OwnerID,
		story.Status,
		story.Priority,
		story.CreatorID,
		story.CreatedAt,
		story.UpdatedAt,
		story.ID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (s *Story) DeleteStory(ct context.Context, tx *transaction.Transaction, storyID uint64) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM story
		WHERE id = $1;
	`, storyID)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewStory(transactionFactory transaction.Factory) *Story {
	return &Story{
		transactionFactory: transactionFactory,
	}
}
