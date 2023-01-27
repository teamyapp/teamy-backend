package dao

import (
	"context"

	"github.com/teamyapp/teamy-backend/core/entity"
)

type User interface {
	FindUserByID(ct context.Context, userID uint64) (entity.User, error)
	FindUsersByIDs(ct context.Context, userIDs []uint64) ([]entity.User, error)
	CreateUser(ct context.Context, user entity.User) error
	UpdateUser(ct context.Context, user entity.User) error
}
