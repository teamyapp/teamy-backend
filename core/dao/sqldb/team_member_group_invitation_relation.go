package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const teamMemberGroupInvitationRelationDaoName = "TeamMemberGroupInvitationRelation"

type TeamMemberGroupInvitationRelation struct {
	metrics dao.Metrics
}

var _ dao.TeamMemberGroupInvitationRelation = TeamMemberGroupInvitationRelation{}

func (t TeamMemberGroupInvitationRelation) FindInvitationIDsByTeamMemberGroupID(
	ct context.Context,
	tx *transaction.Transaction,
	teamMemberGroupID uint64,
) ([]uint64, *errs.Error) {
	t.metrics.ReportDaoOperation(teamMemberGroupInvitationRelationDaoName, "FindInvitationIDsByTeamMemberGroupID")
	query := `
		SELECT
			invitation_id
		FROM team_member_group_invitation_relation
		WHERE group_id = $1;
	`
	rows, err := tx.SQLTx().Query(query, teamMemberGroupID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	var invitationIDs []uint64
	for rows.Next() {
		var invitationID uint64
		err = rows.Scan(
			&invitationID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		invitationIDs = append(invitationIDs, invitationID)
	}

	return invitationIDs, nil
}

func (t TeamMemberGroupInvitationRelation) CreateTeamMemberGroupInvitationRelation(
	ct context.Context,
	tx *transaction.Transaction,
	relation entity.TeamMemberGroupInvitationRelation,
) *errs.Error {
	t.metrics.ReportDaoOperation(teamMemberGroupInvitationRelationDaoName, "CreateTeamMemberGroupInvitationRelation")
	statement := `
		INSERT INTO team_member_group_invitation_relation
		(
		    group_id, 
		    invitation_id, 
		    created_at,
		)
		VALUES ($1, $2, $3);
	`
	_, err := tx.SQLTx().ExecContext(ct,
		statement,
		relation.GroupID,
		relation.InvitationID,
		relation.CreatedAt,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t TeamMemberGroupInvitationRelation) DeleteTeamMemberGroupInvitationRelation(
	ct context.Context,
	tx *transaction.Transaction,
	relation entity.TeamMemberGroupInvitationRelation,
) *errs.Error {
	t.metrics.ReportDaoOperation(teamMemberGroupInvitationRelationDaoName, "DeleteTeamMemberGroupInvitationRelation")
	statement := `
		DELETE FROM team_member_group_invitation_relation
		WHERE group_id = $1 AND invitation_id = $2;
	`
	_, err := tx.SQLTx().ExecContext(ct,
		statement,
		relation.GroupID,
		relation.InvitationID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t TeamMemberGroupInvitationRelation) DeleteTeamMemberGroupInvitationRelationsByGroupID(
	ct context.Context,
	tx *transaction.Transaction,
	groupID uint64,
) *errs.Error {
	t.metrics.ReportDaoOperation(teamMemberGroupInvitationRelationDaoName, "DeleteTeamMemberGroupInvitationRelationsByGroupID")
	statement := `
		DELETE FROM team_member_group_invitation_relation
		WHERE group_id = $1;
	`
	_, err := tx.SQLTx().ExecContext(ct,
		statement,
		groupID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewTeamMemberGroupInvitationRelation(metrics dao.Metrics) TeamMemberGroupInvitationRelation {
	return TeamMemberGroupInvitationRelation{
		metrics: metrics,
	}
}
