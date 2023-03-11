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

type Sprint struct {
	dataCollector telemetry.DataCollector
}

var _ daov2.Sprint = (*Sprint)(nil)

func (s Sprint) FindSprintByID(ct context.Context, tx *transaction.Transaction, sprintID uint64) (entity.Sprint, *errs.Error) {
	sprint := entity.Sprint{}
	err := tx.SQLTx().QueryRow(`
		SELECT
			id,
			start_at,
			end_at,
			created_at,
			owning_team_id
		FROM sprint
		WHERE id = $1;`,
		sprintID).
		Scan(
			&sprint.ID,
			&sprint.StartAt,
			&sprint.EndAt,
			&sprint.CreatedAt,
			&sprint.OwningTeamID,
		)

	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code: errs.NotFound,
			Message: fmt.Sprintf(
				"sprint not found: sprintID=%v", sprintID),
		}
		s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Sprint{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Sprint{}, internalErr
	}

	return sprint, nil
}

func (s Sprint) FindSprintsByIDs(ct context.Context, tx *transaction.Transaction, sprintIDs []uint64) ([]entity.Sprint, *errs.Error) {
	if len(sprintIDs) == 0 {
		return []entity.Sprint{}, nil
	}

	idsString := toIDsString(sprintIDs)
	query := fmt.Sprintf(`
	SELECT
		id,
		start_at,
		end_at,
		created_at,
		owning_team_id
	FROM sprint
	WHERE id IN (%v);`, idsString)
	rows, err := tx.SQLTx().Query(query)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
	var sprints []entity.Sprint
	for rows.Next() {
		var sprint entity.Sprint
		err = rows.
			Scan(
				&sprint.ID,
				&sprint.StartAt,
				&sprint.EndAt,
				&sprint.CreatedAt,
				&sprint.OwningTeamID,
			)
		if err != nil {
			newInternalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}

			if internalErr == nil {
				internalErr = newInternalErr
			}

			s.dataCollector.Logger.ErrorWithContext(ct, newInternalErr)
			continue
		}

		sprints = append(sprints, sprint)
	}

	return sprints, internalErr
}

func (s Sprint) FindSprintsByTeamID(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.Sprint, *errs.Error) {
	rows, err := tx.SQLTx().Query(
		`
	SELECT
		id,
		start_at,
		end_at,
		created_at,
		owning_team_id
	FROM sprint
	WHERE owning_team_id = $1;
`,
		teamID)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
	var sprints []entity.Sprint
	for rows.Next() {
		var sprint entity.Sprint
		err = rows.
			Scan(
				&sprint.ID,
				&sprint.StartAt,
				&sprint.EndAt,
				&sprint.CreatedAt,
				&sprint.OwningTeamID,
			)
		if err != nil {
			newInternalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}

			if internalErr == nil {
				internalErr = newInternalErr
			}

			s.dataCollector.Logger.ErrorWithContext(ct, newInternalErr)
			continue
		}

		sprints = append(sprints, sprint)
	}

	return sprints, internalErr
}

func (s Sprint) FindAllSprints(ct context.Context, tx *transaction.Transaction) ([]entity.Sprint, *errs.Error) {
	rows, err := tx.SQLTx().Query(`
	SELECT
		id,
		start_at,
		end_at,
		created_at,
		owning_team_id
	FROM sprint;
`)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
	var sprints []entity.Sprint
	for rows.Next() {
		var sprint entity.Sprint
		err = rows.
			Scan(
				&sprint.ID,
				&sprint.StartAt,
				&sprint.EndAt,
				&sprint.CreatedAt,
				&sprint.OwningTeamID,
			)
		if err != nil {
			newInternalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}

			if internalErr == nil {
				internalErr = newInternalErr
			}

			s.dataCollector.Logger.ErrorWithContext(ct, newInternalErr)
			continue
		}

		sprints = append(sprints, sprint)
	}

	return sprints, internalErr
}

func (s Sprint) CreateSprint(ct context.Context, tx *transaction.Transaction, sprint entity.Sprint) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		INSERT INTO sprint
		(
			id,
			start_at,
			end_at,
			created_at,
			owning_team_id
		)
		VALUES ($1, $2, $3, $4, $5);`,
		sprint.ID,
		sprint.StartAt,
		sprint.EndAt,
		sprint.CreatedAt,
		sprint.OwningTeamID,
	)

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (s Sprint) DeleteSprint(ct context.Context, tx *transaction.Transaction, sprintID uint64) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		DELETE FROM sprint
		WHERE id = $1;
		`,
		sprintID)

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func NewSprint(dataCollector telemetry.DataCollector) Sprint {
	return Sprint{dataCollector: dataCollector}
}
