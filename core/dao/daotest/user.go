package daotest

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type User struct {
	db *dbtest.InMemoryDB
}

var _ dao.User = (*User)(nil)

func (u User) FindUserByID(ct context.Context, userID uint64) (entity.User, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (u User) FindUsersByIDs(ct context.Context, userIDs []uint64) ([]entity.User, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (u User) CreateUser(ct context.Context, user entity.User) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (u User) UpdateUser(ct context.Context, user entity.User) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewUser(db *dbtest.InMemoryDB) User {
	return User{db: db}
}
