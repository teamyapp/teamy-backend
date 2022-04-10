package dao

import "github.com/teamyapp/teamy-backend/app/entityv2"

type User interface {
	FindUserByID(id uint64) (entityv2.User, error)
}
