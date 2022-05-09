package dao

import "github.com/teamyapp/teamy-backend/app/entityv2"

type Thread interface {
	FindThreadByID(threadID uint64) (entityv2.Thread, error)
}
