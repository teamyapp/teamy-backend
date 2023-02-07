package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type User interface {
	FindUserByID(ct context.Context, userID uint64) (entity.User, *errs.Error)
	FindUsersByIDs(ct context.Context, userIDs []uint64) ([]entity.User, *errs.Error)
	CreateUser(ct context.Context, user entity.User) *errs.Error
	UpdateUser(ct context.Context, user entity.User) *errs.Error
}
