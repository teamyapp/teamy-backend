package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/teamyapp/teamy-backend/app/dao"
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type Invitation struct {
	db *sql.DB
}

var _ dao.Invitation = (*Invitation)(nil)

func (i Invitation) FindInvitationByID(id uint64) (entityv2.Invitation, error) {
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
		create_at
	FROM invitation
	WHERE id = $1;
`
	invitation := entityv2.Invitation{}
	err := i.db.QueryRow(statement, id).
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
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entityv2.Invitation{}, dao.ErrNotFound(fmt.Sprintf(
			"invitation not found: id=%v",
			id))
	}

	return invitation, err
}

func (i Invitation) FindInvitationsByTeamID(teamID uint64) ([]entityv2.Invitation, error) {
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
		create_at
	FROM invitation
	WHERE team_id = $1;
`
	rows, err := i.db.Query(statement, teamID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	invitations := make([]entityv2.Invitation, 0)
	for rows.Next() {
		invitation := entityv2.Invitation{}
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
		)
		if err != nil {
			log.Println(err)
		}

		invitations = append(invitations, invitation)
	}

	return invitations, err
}

func (i Invitation) FindAllInvitations() ([]entityv2.Invitation, error) {
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
		create_at
	FROM invitation;
`
	rows, err := i.db.Query(statement)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	invitations := make([]entityv2.Invitation, 0)
	for rows.Next() {
		invitation := entityv2.Invitation{}
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
		)
		if err != nil {
			log.Println(err)
		}

		invitations = append(invitations, invitation)
	}

	return invitations, err
}

func (i Invitation) CreateInvitation(invitation entityv2.Invitation) error {
	statement := `
	INSERT INTO invitation
	(
	    id,
		sender_user_id,
		receiver_first_name,
		receiver_email,
	    team_id,
		expire_at,
		status,
		code,
		create_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
`
	_, err := i.db.Exec(statement,
		int64(invitation.ID),
		invitation.SenderUserID,
		invitation.ReceiverFirstName,
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

func (i Invitation) UpdateInvitation(invitation entityv2.Invitation) error {
	//TODO implement me
	panic("implement me")
}

func (i Invitation) DeleteInvitation(id uint64) error {
	//TODO implement me
	panic("implement me")
}

func NewInvitation(sqlDB *sql.DB) Invitation {
	return Invitation{db: sqlDB}
}
