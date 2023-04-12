package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Sprint struct {
	transactionFactory transaction.Factory
}

var _ daov2.Sprint = (*Sprint)(nil)

func (s Sprint) FindSprintByID(ct context.Context, sprintID uint64) (entity.Sprint, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := s.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.Sprint{}, err
	}

	defer tx.Rollback()
	return s.FindSprintByIDWithTx(ct, tx, sprintID)
}

func (s Sprint) FindSprintsByTeamID(ct context.Context, teamID uint64) ([]entity.Sprint, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := s.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return s.FindSprintsByTeamIDWithTx(ct, tx, teamID)
}

func (s Sprint) FindAllSprints(ct context.Context) ([]entity.Sprint, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := s.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return s.FindAllSprintsWithTx(ct, tx)
}

func (s Sprint) FindSprintByIDWithTx(ct context.Context, tx *transaction.Transaction, sprintID uint64) (entity.Sprint, *errs.Error) {
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
		return entity.Sprint{}, errs.NewError(
			errs.NotFound,
			fmt.Sprintf(
				"sprint not found: sprintID=%v", sprintID))
	}

	if err != nil {
		return entity.Sprint{}, errs.NewError(errs.Unknown, err.Error())
	}

	return sprint, nil
}

func (s Sprint) FindSprintsByIDsWithTx(ct context.Context, tx *transaction.Transaction, sprintIDs []uint64) ([]entity.Sprint, *errs.Error) {
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
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

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
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		sprints = append(sprints, sprint)
	}

	return sprints, nil
}

func (s Sprint) FindSprintsByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.Sprint, *errs.Error) {
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
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

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
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		sprints = append(sprints, sprint)
	}

	return sprints, nil
}

func (s Sprint) FindAllSprintsWithTx(ct context.Context, tx *transaction.Transaction) ([]entity.Sprint, *errs.Error) {
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
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

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
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		sprints = append(sprints, sprint)
	}

	return sprints, nil
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
		return errs.NewError(errs.Unknown, err.Error())
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
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewSprint(transactionFactory transaction.Factory) Sprint {
	return Sprint{
		transactionFactory: transactionFactory,
	}
}
