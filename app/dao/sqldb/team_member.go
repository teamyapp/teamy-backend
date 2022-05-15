package sqldb

import (
	"database/sql"

	"github.com/teamyapp/teamy-backend/app/dao"
)

type TeamMember struct {
	db *sql.DB
}

var _ dao.TeamMember = (*TeamMember)(nil)

func (t TeamMember) FindTeamIDsByUserID(userID uint64) ([]uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (t TeamMember) FindTeamMemberIDsByTeamID(teamID uint64) ([]uint64, error) {
	//TODO implement me
	panic("implement me")
}

func NewTeamMember(sqlDB *sql.DB) TeamMember {
	return TeamMember{db: sqlDB}
}
