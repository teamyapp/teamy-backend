package sqldb

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Phase struct {
	transactionFactory transaction.Factory
}

var _ dao.Phase = (*Phase)(nil)

func (p *Phase) FindPhaseByIDWithTx(ct context.Context, tx *transaction.Transaction, phaseID uint64) (entity.Phase, *errs.Error) {
	phase := entity.Phase{}
	err := tx.SQLTx().QueryRowContext(ct, `
		SELECT
			id,
			name,
			status,
			expected_start_at,
			actual_start_at,
			expected_end_at,
			actual_end_at,
			creator_id,
			created_at,
			updated_at
		FROM phase
		WHERE id = $1
	`, phaseID).Scan(
		&phase.ID,
		&phase.Name,
		&phase.Status,
		&phase.ExpectedStartAt,
		&phase.ActualStartAt,
		&phase.ExpectedEndAt,
		&phase.ActualEndAt,
		&phase.CreatorID,
		&phase.CreatedAt,
		&phase.UpdatedAt,
	)

	if err != nil {
		return entity.Phase{}, errs.NewError(errs.Unknown, err.Error())
	}

	return phase, nil
}

func (p *Phase) FindPhasesByIDsWithTx(ct context.Context, tx *transaction.Transaction, phaseIDs []uint64) ([]entity.Phase, *errs.Error) {
	phases := []entity.Phase{}
	idsStr := toIDsString(phaseIDs)
	query := fmt.Sprintf(`
		SELECT
			id,
			name,
			status,
			expected_start_at,
			actual_start_at,
			expected_end_at,
			actual_end_at,
			creator_id,
			created_at,
			updated_at
		FROM phase
		WHERE id IN (%s)
	`, idsStr)

	rows, err := tx.SQLTx().QueryContext(ct, query)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		var phase entity.Phase
		err := rows.Scan(
			&phase.ID,
			&phase.Name,
			&phase.Status,
			&phase.ExpectedStartAt,
			&phase.ActualStartAt,
			&phase.ExpectedEndAt,
			&phase.ActualEndAt,
			&phase.CreatorID,
			&phase.CreatedAt,
			&phase.UpdatedAt,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}
		phases = append(phases, phase)
	}

	return phases, nil
}

func (p *Phase) CreatePhase(ct context.Context, tx *transaction.Transaction, phase entity.Phase) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		INSERT INTO phase (
			id,
			name,
			status,
			expected_start_at,
			actual_start_at,
			expected_end_at,
			actual_end_at,
			creator_id,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
	`,
		phase.ID,
		phase.Name,
		phase.Status,
		phase.ExpectedStartAt,
		phase.ActualStartAt,
		phase.ExpectedEndAt,
		phase.ActualEndAt,
		phase.CreatorID,
		phase.CreatedAt,
		phase.UpdatedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (p *Phase) UpdatePhase(ct context.Context, tx *transaction.Transaction, phase entity.Phase) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		UPDATE phase
		SET
			name = $1,
			status = $2,
			expected_start_at = $3,
			actual_start_at = $4,
			expected_end_at = $5,
			actual_end_at = $6,
			created_at = $7,
			updated_at = $8
		WHERE id = $9;
	`,
		phase.Name,
		phase.Status,
		phase.ExpectedStartAt,
		phase.ActualStartAt,
		phase.ExpectedEndAt,
		phase.ActualEndAt,
		phase.CreatedAt,
		phase.UpdatedAt,
		phase.ID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (p *Phase) DeletePhase(ct context.Context, tx *transaction.Transaction, phaseID uint64) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM phase
		WHERE id = $1;
	`, phaseID)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewPhase(transactionFactory transaction.Factory) *Phase {
	return &Phase{
		transactionFactory: transactionFactory,
	}
}
