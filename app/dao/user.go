package dao

import (
	"github.com/teamyapp/teamy-backend/app/entity"
)

type User interface {
	FindUserByID(userID uint64) (entity.User, error)
	FindUsersByIDs(userIDs []uint64) ([]entity.User, error)
	CreateUser(user entity.User) error
	UpdateUser(user entity.User) error
}
