package daotestv2

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamMember struct {
	db *dbtest.InMemoryDB
}

var _ daov2.TeamMember = (*TeamMember)(nil)

func (t TeamMember) FindTeamIDsByUserID(ct context.Context, tx *transaction.Transaction, userID uint64) ([]uint64, *errs.Error) {
	var teamIDs []uint64
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currTeamMember := rawRow.(entity.TeamMember)
				if currTeamMember.UserID == userID {
					teamIDs = append(teamIDs, currTeamMember.TeamID)
					return nil
				}
			}

			return nil
		},
	})
	return teamIDs, err
}

func (t TeamMember) FindTeamMemberIDsByTeamID(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]uint64, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t TeamMember) FindTeamMembersByTeamID(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.TeamMember, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t TeamMember) FindTeamMember(ct context.Context, tx *transaction.Transaction, teamID uint64, userID uint64) (entity.TeamMember, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t TeamMember) CreateTeamMember(ct context.Context, tx *transaction.Transaction, teamMember entity.TeamMember) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (t TeamMember) UpdateTeamMember(ct context.Context, tx *transaction.Transaction, teamMember entity.TeamMember) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (t TeamMember) DeleteTeamMember(ct context.Context, tx *transaction.Transaction, teamID uint64, userID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewTeamMember(db *dbtest.InMemoryDB) TeamMember {
	return TeamMember{db: db}
}
