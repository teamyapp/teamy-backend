package dao

import (
	"log"

	"database/sql"

	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type Invitation interface {
	FindInvitationByID(id uint64) (entityv2.Invitation, error)
	FindInvitationsByTeamID(teamID uint64) ([]entityv2.Invitation, error)
}

type SQLInvitation struct {
	db *sql.DB
}

var _ Invitation = (*SQLInvitation)(nil)

func (S SQLInvitation) FindInvitationByID(id uint64) (entityv2.Invitation, error) {
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
	err := S.db.QueryRow(statement, int(id)).
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

	if err != nil {
		log.Println(err)
	}

	return invitation, err
}

func (S SQLInvitation) FindInvitationsByTeamID(teamID uint64) ([]entityv2.Invitation, error) {
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
	rows, err := S.db.Query(statement, int(teamID))
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
