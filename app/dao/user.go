package dao

import "github.com/teamyapp/teamy-backend/app/entityv2"

type User interface {
	FindUser(id uint64) (entityv2.User, error)
}
