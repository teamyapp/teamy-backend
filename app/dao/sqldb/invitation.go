package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/teamyapp/teamy-backend/app/dao"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Invitation struct {
	db *sql.DB
}

var _ dao.Invitation = (*Invitation)(nil)

func (i Invitation) FindInvitationByID(invitationID uint64) (entity.Invitation, error) {
	statement := `
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
`
	invitation := entity.Invitation{}
	err := i.db.QueryRow(statement, invitationID).
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

	if errors.Is(err, sql.ErrNoRows) {
		return entity.Invitation{}, dao.ErrNotFound(fmt.Sprintf(
			"invitation not found: id=%v", invitationID))
	}

	return invitation, err
}

func (i Invitation) FindInvitationsByTeamID(teamID uint64) ([]entity.Invitation, error) {
	statement := `
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
`
	rows, err := i.db.Query(statement, teamID)
	if err != nil {
		log.Println(err)
		return nil, err
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
			log.Println(err)
			continue
		}

		invitations = append(invitations, invitation)
	}

	return invitations, err
}

func (i Invitation) FindAllInvitations() ([]entity.Invitation, error) {
	statement := `
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
`
	rows, err := i.db.Query(statement)
	if err != nil {
		log.Println(err)
		return nil, err
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
			log.Println(err)
			continue
		}

		invitations = append(invitations, invitation)
	}

	return invitations, err
}

func (i Invitation) CreateInvitation(invitation entity.Invitation) error {
	statement := `
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
`
	_, err := i.db.Exec(statement,
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
		log.Println(err)
	}

	return err
}

func (i Invitation) UpdateInvitation(invitation entity.Invitation) error {
	_, err := i.db.Exec(`
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
	return err
}

func (i Invitation) DeleteInvitation(invitationID uint64) error {
	_, err := i.db.Exec(`
		DELETE FROM invitation
		WHERE id = $1;
		`,
		invitationID)
	return err
}

func NewInvitation(sqlDB *sql.DB) Invitation {
	return Invitation{db: sqlDB}
}
