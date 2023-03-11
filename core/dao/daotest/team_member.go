package daotest

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamMember struct {
	db *dbtest.InMemoryDB
}

var _ dao.TeamMember = (*TeamMember)(nil)

func (t TeamMember) FindTeamIDsByUserID(ct context.Context, userID uint64) ([]uint64, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t TeamMember) FindTeamMemberIDsByTeamID(ct context.Context, teamID uint64) ([]uint64, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t TeamMember) FindTeamMembersByTeamID(ct context.Context, teamID uint64) ([]entity.TeamMember, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t TeamMember) FindTeamMember(ct context.Context, teamID uint64, userID uint64) (entity.TeamMember, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t TeamMember) CreateTeamMember(ct context.Context, teamMember entity.TeamMember) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (t TeamMember) UpdateTeamMember(ct context.Context, teamMember entity.TeamMember) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (t TeamMember) DeleteTeamMember(ct context.Context, teamID uint64, userID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewTeamMember(db *dbtest.InMemoryDB) TeamMember {
	return TeamMember{db: db}
}
