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

type Invitation struct {
	logger             telemetry.Logger
	transactionFactory transaction.Factory
}

var _ daov2.Invitation = (*Invitation)(nil)

func (i Invitation) FindInvitationByID(ct context.Context, invitationID uint64) (entity.Invitation, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := i.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.Invitation{}, err
	}

	defer tx.Rollback()
	return i.FindInvitationByIDWithTx(ct, tx, invitationID)
}

func (i Invitation) FindInvitationsByTeamID(ct context.Context, teamID uint64) ([]entity.Invitation, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := i.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return i.FindInvitationsByTeamIDWithTx(ct, tx, teamID)
}

func (i Invitation) FindAllInvitations(ct context.Context) ([]entity.Invitation, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := i.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return i.FindAllInvitationsWithTx(ct, tx)
}

func (i Invitation) FindInvitationByIDWithTx(ct context.Context, tx *transaction.Transaction, invitationID uint64) (entity.Invitation, *errs.Error) {
	invitation := entity.Invitation{}
	err := tx.SQLTx().QueryRow(`
	SELECT
		id,
		sender_user_id,
		receiver_first_name,
		receiver_last_name,
		receiver_user_id,
		receiver_email,
		team_id,
		expire_at,
		status,
		code,
		created_at,
		updated_at
	FROM invitation
	WHERE id = $1;
`,
		invitationID).
		Scan(
			&invitation.ID,
			&invitation.SenderUserID,
			&invitation.ReceiverFirstName,
			&invitation.ReceiverLastName,
			&invitation.ReceiverUserID,
			&invitation.ReceiverEmail,
			&invitation.TeamID,
			&invitation.ExpireAt,
			&invitation.Status,
			&invitation.Code,
			&invitation.CreatedAt,
			&invitation.UpdatedAt,
		)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Invitation{}, errs.NewError(
			    errs.NotFound, 
			    fmt.Sprintf("invitation not found: invitationID=%v", invitationID))
		}

		return entity.Invitation{}, errs.NewError(errs.Unknown, err.Error())
	}

	return invitation, nil
}

func (i Invitation) FindInvitationsByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.Invitation, *errs.Error) {
	rows, err := tx.SQLTx().Query(`
	SELECT
		id,
		sender_user_id,
		receiver_first_name,
		receiver_last_name,
		receiver_user_id,
		receiver_email,
		team_id,
		expire_at,
		status,
		code,
		created_at,
		updated_at
	FROM invitation
	WHERE team_id = $1;
`,
		teamID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	invitations := make([]entity.Invitation, 0)
	for rows.Next() {
		invitation := entity.Invitation{}
		err = rows.Scan(
			&invitation.ID,
			&invitation.SenderUserID,
			&invitation.ReceiverFirstName,
			&invitation.ReceiverLastName,
			&invitation.ReceiverUserID,
			&invitation.ReceiverEmail,
			&invitation.TeamID,
			&invitation.ExpireAt,
			&invitation.Status,
			&invitation.Code,
			&invitation.CreatedAt,
			&invitation.UpdatedAt,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		invitations = append(invitations, invitation)
	}

	return invitations, nil
}

func (i Invitation) FindAllInvitationsWithTx(ct context.Context, tx *transaction.Transaction) ([]entity.Invitation, *errs.Error) {
	rows, err := tx.SQLTx().Query(`
	SELECT
		id,
		sender_user_id,
		receiver_first_name,
		receiver_last_name,
		receiver_user_id,
		receiver_email,
		team_id,
		expire_at,
		status,
		code,
		created_at,
		updated_at
	FROM invitation;
`)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	invitations := make([]entity.Invitation, 0)
	for rows.Next() {
		invitation := entity.Invitation{}
		err = rows.Scan(
			&invitation.ID,
			&invitation.SenderUserID,
			&invitation.ReceiverFirstName,
			&invitation.ReceiverLastName,
			&invitation.ReceiverUserID,
			&invitation.ReceiverEmail,
			&invitation.TeamID,
			&invitation.ExpireAt,
			&invitation.Status,
			&invitation.Code,
			&invitation.CreatedAt,
			&invitation.UpdatedAt,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		invitations = append(invitations, invitation)
	}

	return invitations, nil
}

func (i Invitation) CreateInvitation(ct context.Context, tx *transaction.Transaction, invitation entity.Invitation) *errs.Error {
	_, err := tx.SQLTx().Exec(`
	INSERT INTO invitation
	(
	    id,
		sender_user_id,
		receiver_first_name,
	 	receiver_last_name,
		receiver_email,
	    team_id,
		expire_at,
		status,
		code,
		created_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
`,
		int64(invitation.ID),
		invitation.SenderUserID,
		invitation.ReceiverFirstName,
		invitation.ReceiverLastName,
		invitation.ReceiverEmail,
		invitation.TeamID,
		invitation.ExpireAt,
		invitation.Status,
		invitation.Code,
		invitation.CreatedAt,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (i Invitation) UpdateInvitation(ct context.Context, tx *transaction.Transaction, invitation entity.Invitation) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		UPDATE invitation
		SET
			receiver_first_name = $1,
			receiver_last_name = $2,
			receiver_user_id = $3,
			expire_at = $4,
			status = $5,
			code = $6,
			updated_at = $7
		WHERE id = $8;`,
		invitation.ReceiverFirstName,
		invitation.ReceiverLastName,
		invitation.ReceiverUserID,
		invitation.ExpireAt,
		invitation.Status,
		invitation.Code,
		invitation.UpdatedAt,
		invitation.ID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (i Invitation) DeleteInvitation(ct context.Context, tx *transaction.Transaction, invitationID uint64) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		DELETE FROM invitation
		WHERE id = $1;
		`,
		invitationID)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewInvitation(logger telemetry.Logger, transactionFactory transaction.Factory) Invitation {
	return Invitation{logger: logger, transactionFactory: transactionFactory}
}
