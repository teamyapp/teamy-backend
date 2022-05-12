package repo

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/teamyapp/teamy-backend/app/entity"
)

type Team interface {
	FindActiveTeam(userID uint64) (*entity.Team, error)
	FindAllTeamIDs(userID uint64) ([]uint64, error)
	FindTeams(teamIDs []uint64) ([]entity.Team, error)
	ListTeamMemberIDs(teamID uint64) ([]uint64, error)
}

type SQLTeam struct {
	db *sql.DB
}

var _ Team = (*SQLTeam)(nil)

func (S SQLTeam) ListTeamMemberIDs(teamID uint64) ([]uint64, error) {
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

	var ids []uint64
	var id uint64
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

func (S SQLTeam) FindActiveTeam(userID uint64) (*entity.Team, error) {
	team := entity.Team{}
	err := S.db.
		QueryRow(`
SELECT team.id, team.name, team.logo_url, team.created_at, team.updated_at
FROM user_state
INNER JOIN team ON user_state.active_team_id = team.id
WHERE user_id = $1`,
			int(userID)).
		Scan(&team.ID, &team.Name, &team.IconURL, &team.CreatedAt, &team.UpdatedAt)
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

func (S SQLTeam) FindAllTeamIDs(userID uint64) ([]uint64, error) {
	rows, err := S.db.Query(`
	SELECT team_id
	FROM team_member
	WHERE user_id = $1;
`, int(userID))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	var ids []uint64
	var id uint64
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

func (S SQLTeam) FindTeams(teamIDs []uint64) ([]entity.Team, error) {
	idsString := toIDsString(teamIDs)
	query := fmt.Sprintf(`
SELECT id, name, logo_url, created_at, updated_at
FROM team
WHERE id IN (%s);`, idsString)
	rows, err := S.db.Query(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	var teams []entity.Team
	for rows.Next() {
		var team entity.Team
		err = rows.Scan(&team.ID, &team.Name, &team.IconURL, &team.CreatedAt, &team.UpdatedAt)
		if err != nil {
			log.Println(team.ID, err)
			continue
		}
		teams = append(teams, team)
	}

	return teams, nil
}

func NewSQLTeam(db *sql.DB) SQLTeam {
	return SQLTeam{db: db}
}
