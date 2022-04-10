package dao

import "github.com/teamyapp/teamy-backend/app/entityv2"

type Team interface {
	FindTeam(id uint64) (entityv2.Team, error)
}
