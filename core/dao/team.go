package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Team interface {
	FindTeamByID(ct context.Context, teamID uint64) (entity.Team, *errs.Error)
	FindAllTeams(ct context.Context) ([]entity.Team, *errs.Error)
	FindAllTeamsWithTx(ct context.Context, tx *transaction.Transaction) ([]entity.Team, *errs.Error)
	FindTeamByIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) (entity.Team, *errs.Error)
	FindTeamsByIDsWithTx(ct context.Context, tx *transaction.Transaction, teamIDs []uint64) ([]entity.Team, *errs.Error)
	CreateTeam(ct context.Context, tx *transaction.Transaction, team entity.Team) *errs.Error
	UpdateTeam(ct context.Context, tx *transaction.Transaction, team entity.Team) *errs.Error
	DeleteTeam(ct context.Context, tx *transaction.Transaction, teamID uint64) *errs.Error
}
