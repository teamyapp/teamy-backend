package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type PhaseStoryRelation struct {
	transactionFactory transaction.Factory
}

var _ dao.PhaseStoryRelation = (*PhaseStoryRelation)(nil)

func (p *PhaseStoryRelation) FindStoryIDsByPhaseIDWithTx(ct context.Context, tx *transaction.Transaction, phaseID uint64) ([]uint64, *errs.Error) {
	storyIDs := []uint64{}
	rows, err := tx.SQLTx().QueryContext(ct, `
		SELECT
			story_id
		FROM phase_story_relation
		WHERE phase_id = $1
	`, phaseID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		var storyID uint64
		err := rows.Scan(&storyID)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		storyIDs = append(storyIDs, storyID)
	}

	return storyIDs, nil
}

func (p *PhaseStoryRelation) CreatePhaseStoryRelation(ct context.Context, tx *transaction.Transaction, phaseStroyRelation entity.PhaseStoryRelation) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		INSERT INTO phase_story_relation (
			phase_id,
			story_id
		) VALUES (
			$1,
			$2
		)
	`,
		phaseStroyRelation.PhaseID,
		phaseStroyRelation.StoryID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (p *PhaseStoryRelation) DeletePhaseStoryRelation(ct context.Context, tx *transaction.Transaction, phaseID uint64, storyID uint64) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM phase_story_relation
		WHERE phase_id = $1 AND story_id = $2
	`,
		phaseID,
		storyID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (p *PhaseStoryRelation) DeletePhaseStoryRelationByPhaseID(ct context.Context, tx *transaction.Transaction, phaseID uint64) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM phase_story_relation
		WHERE phase_id = $1
	`, phaseID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewPhaseStoryRelation(transactionFactory transaction.Factory) *PhaseStoryRelation {
	return &PhaseStoryRelation{
		transactionFactory: transactionFactory,
	}
}
