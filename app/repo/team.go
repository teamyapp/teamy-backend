package repo

import (
	"database/sql"
	"log"

	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
	"github.com/teamyapp/teamy-backend/app/errs"
)

type Team interface {
	GetActiveTeam(userID oneEntity.ID) (entity.Team, error)
}

type SQLTeam struct {
	db *sql.DB
}

var _ Team = (*SQLTeam)(nil)

func (S SQLTeam) GetActiveTeam(userID oneEntity.ID) (entity.Team, error) {
	team := entity.Team{}
	err := S.db.
		QueryRow(`
SELECT team.id, team.name, team.logo_url, team.created_at, team.updated_at
FROM user_state
INNER JOIN team ON user_state.active_team_id = team.id
WHERE user_id = $1`,
			int(userID)).
		Scan(&team.ID, &team.Name, &team.LogoURL, &team.CreatedAt, &team.UpdatedAt)
	if err != nil {
		log.Println(err)
		if err == sql.ErrNoRows {
			return entity.Team{}, errs.NoActiveTeam(userID)
		} else {
			return entity.Team{}, err
		}
	}
	return team, err
}

func NewSQLTeam(db *sql.DB) SQLTeam {
	return SQLTeam{db: db}
}
