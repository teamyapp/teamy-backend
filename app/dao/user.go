package dao

import "github.com/teamyapp/teamy-backend/app/entityv2"

type User interface {
	FindUserByID(userID uint64) (entityv2.User, error)
	FindUsersByIDs(userIDs []uint64) ([]entityv2.User, error)
	CreateUser(user entityv2.User) error
}
