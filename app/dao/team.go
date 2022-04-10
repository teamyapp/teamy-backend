package dao

import "github.com/teamyapp/teamy-backend/app/entityv2"

type Team interface {
	FindTeamByID(id uint64) (entityv2.Team, error)
	FindTeamsByIDs(ids []uint64) ([]entityv2.Team, error)
}
