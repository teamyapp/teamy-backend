package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintParticipant struct {
	logger             telemetry.Logger
	transactionFactory transaction.Factory
}

var _ dao.SprintParticipant = (*SprintParticipant)(nil)

func (s SprintParticipant) FindParticipantIDsBySprintID(ct context.Context, sprintID uint64) ([]uint64, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := s.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return s.FindParticipantIDsBySprintIDWithTx(ct, tx, sprintID)
}

func (s SprintParticipant) FindParticipantsBySprintID(ct context.Context, sprintID uint64) ([]entity.SprintParticipant, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := s.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return s.FindParticipantsBySprintIDWithTx(ct, tx, sprintID)
}

func (s SprintParticipant) FindParticipant(ct context.Context, sprintID uint64, participantUserID uint64) (entity.SprintParticipant, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := s.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.SprintParticipant{}, err
	}

	defer tx.Rollback()
	return s.FindParticipantWithTx(ct, tx, sprintID, participantUserID)
}

func (s SprintParticipant) FindParticipantIDsBySprintIDWithTx(ct context.Context, tx *transaction.Transaction, sprintID uint64) ([]uint64, *errs.Error) {
	rows, err := tx.SQLTx().Query(
		`
	SELECT
		user_id
	FROM sprint_participant
	WHERE sprint_id = $1;
`,
		sprintID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	participantUserIDs := make([]uint64, 0)
	for rows.Next() {
		var participantUserID uint64
		err = rows.Scan(
			&participantUserID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		participantUserIDs = append(participantUserIDs, participantUserID)
	}

	return participantUserIDs, nil
}

func (s SprintParticipant) FindParticipantsBySprintIDWithTx(ct context.Context, tx *transaction.Transaction, sprintID uint64) ([]entity.SprintParticipant, *errs.Error) {
	rows, err := tx.SQLTx().Query(
		`
	SELECT
		sprint_id,
		user_id,
		total_bandwidth,
	 	unused_bandwidth,
		created_at,
	 	updated_at
	FROM sprint_participant
	WHERE sprint_id = $1;
`,
		sprintID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	sprintParticipants := make([]entity.SprintParticipant, 0)
	for rows.Next() {
		sprintParticipant := entity.SprintParticipant{}
		err = rows.Scan(
			&sprintParticipant.SprintID,
			&sprintParticipant.UserID,
			&sprintParticipant.TotalBandwidth,
			&sprintParticipant.UnusedBandwidth,
			&sprintParticipant.CreatedAt,
			&sprintParticipant.UpdatedAt,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		sprintParticipants = append(sprintParticipants, sprintParticipant)
	}

	return sprintParticipants, nil
}

func (s SprintParticipant) FindParticipantWithTx(ct context.Context, tx *transaction.Transaction, sprintID uint64, participantUserID uint64) (entity.SprintParticipant, *errs.Error) {
	participant := entity.SprintParticipant{}
	err := tx.SQLTx().QueryRow(`
	SELECT
		sprint_id,
		user_id,
		total_bandwidth,
		unused_bandwidth,
		created_at,
		updated_at
	FROM sprint_participant
	WHERE sprint_id = $1 AND user_id=$2;
`,
		sprintID,
		participantUserID).
		Scan(
			&participant.SprintID,
			&participant.UserID,
			&participant.TotalBandwidth,
			&participant.UnusedBandwidth,
			&participant.CreatedAt,
			&participant.UpdatedAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.SprintParticipant{}, errs.NewError(errs.NotFound, fmt.Sprintf(
			"participant not found: sprintID=%v, participantUserID=%v", sprintID, participantUserID))
	}

	if err != nil {
		return entity.SprintParticipant{}, errs.NewError(errs.Unknown, err.Error())
	}

	return participant, nil
}

func (s SprintParticipant) CreateSprintParticipant(ct context.Context, tx *transaction.Transaction, participant entity.SprintParticipant) *errs.Error {
	_, err := tx.SQLTx().Exec(`
	INSERT INTO sprint_participant
	(
	    sprint_id,
		user_id,
		total_bandwidth,
	 	unused_bandwidth,
		created_at,
	 	updated_at
	)
	VALUES ($1, $2, $3, $4, $5, $6);
`,
		participant.SprintID,
		participant.UserID,
		participant.TotalBandwidth,
		participant.UnusedBandwidth,
		participant.CreatedAt,
		participant.UpdatedAt,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (s SprintParticipant) UpdateSprintParticipant(ct context.Context, tx *transaction.Transaction, participant entity.SprintParticipant) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		UPDATE sprint_participant
		SET
		    sprint_id = $1,
			user_id = $2,
			total_bandwidth = $3,
			unused_bandwidth = $4,
			created_at = $5,
			updated_at = $6
		WHERE sprint_id = $7 AND user_id = $8;`,
		participant.SprintID,
		participant.UserID,
		participant.TotalBandwidth,
		participant.UnusedBandwidth,
		participant.CreatedAt,
		participant.UpdatedAt,
		participant.SprintID,
		participant.UserID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (s SprintParticipant) DeleteSprintParticipant(ct context.Context, tx *transaction.Transaction, sprintID uint64, userID uint64) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		DELETE FROM sprint_participant
		WHERE sprint_id = $1 AND user_id = $2;
		`,
		sprintID,
		userID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewSprintParticipant(logger telemetry.Logger, transactionFactory transaction.Factory) SprintParticipant {
	return SprintParticipant{
		logger:             logger,
		transactionFactory: transactionFactory,
	}
}
