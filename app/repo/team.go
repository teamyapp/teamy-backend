package repo

import (
	"database/sql"
	"log"

	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Team interface {
	GetActiveTeam(userID oneEntity.ID) (*entity.Team, error)
	ListTeamMemberIDs(teamID oneEntity.ID) ([]oneEntity.ID, error)
}

type SQLTeam struct {
	db *sql.DB
}

var _ Team = (*SQLTeam)(nil)

func (S SQLTeam) ListTeamMemberIDs(teamID oneEntity.ID) ([]oneEntity.ID, error) {
	rows, err := S.db.Query(`
	SELECT user_id
	FROM team_member
	WHERE team_id = $1
`, int(teamID))

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	var ids []oneEntity.ID
	var id oneEntity.ID
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			log.Println(id, err)
			continue
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (S SQLTeam) GetActiveTeam(userID oneEntity.ID) (*entity.Team, error) {
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
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			log.Println(err)
			return nil, err
		}
	}
	return &team, err
}

func NewSQLTeam(db *sql.DB) SQLTeam {
	return SQLTeam{db: db}
}
