package daotest

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Team struct {
	db *dbtest.InMemoryDB
}

var _ dao.Team = (*Team)(nil)

func (t Team) FindAllTeams(ct context.Context) ([]entity.Team, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t Team) FindTeamByID(ct context.Context, teamID uint64) (entity.Team, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t Team) FindTeamsByIDs(ct context.Context, teamIDs []uint64) ([]entity.Team, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t Team) CreateTeam(ct context.Context, team entity.Team) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (t Team) UpdateTeam(ct context.Context, team entity.Team) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (t Team) DeleteTeam(ct context.Context, teamID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewTeam(db *dbtest.InMemoryDB) Team {
	return Team{db: db}
}
