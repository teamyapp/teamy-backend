package sqldb

import (
	"database/sql"

	"github.com/teamyapp/teamy-backend/app/dao"
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type Team struct {
	db *sql.DB
}

var _ dao.Team = (*Team)(nil)

func (t Team) FindTeamByID(id uint64) (entityv2.Team, error) {
	//TODO implement me
	panic("implement me")
}

func (t Team) FindTeamsByIDs(ids []uint64) ([]entityv2.Team, error) {
	//TODO implement me
	panic("implement me")
}

func NewTeam(sqlDB *sql.DB) Team {
	return Team{db: sqlDB}
}
