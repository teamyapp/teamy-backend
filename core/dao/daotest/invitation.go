package daotest

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Invitation struct {
	db *dbtest.InMemoryDB
}

var _ dao.Invitation = (*Invitation)(nil)

func (i Invitation) FindAllInvitations(ct context.Context) ([]entity.Invitation, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (i Invitation) FindInvitationByID(ct context.Context, invitationID uint64) (entity.Invitation, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (i Invitation) FindInvitationsByTeamID(ct context.Context, teamID uint64) ([]entity.Invitation, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (i Invitation) CreateInvitation(ct context.Context, invitation entity.Invitation) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (i Invitation) UpdateInvitation(ct context.Context, invitation entity.Invitation) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (i Invitation) DeleteInvitation(ct context.Context, invitationID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewInvitation(db *dbtest.InMemoryDB) Invitation {
	return Invitation{db: db}
}
