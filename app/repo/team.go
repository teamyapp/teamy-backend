package repo

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"strconv"

	"log"

	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Team interface {
	FindActiveTeam(userID oneEntity.ID) (*entity.Team, error)
	FindAllTeamIDs(userID oneEntity.ID) ([]oneEntity.ID, error)
	FindTeams(teamIDs []oneEntity.ID) ([]entity.Team, error)
	ListTeamMemberIDs(teamID oneEntity.ID) ([]oneEntity.ID, error)
	AddUserToTeam(userID oneEntity.ID, teamID oneEntity.ID) (bool, error)
	TeamExists(teamID oneEntity.ID) (bool, error)
	IsUserInTeam(userID oneEntity.ID, teamID oneEntity.ID) (bool, error)
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

func (S SQLTeam) FindActiveTeam(userID oneEntity.ID) (*entity.Team, error) {
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

func (S SQLTeam) FindAllTeamIDs(userID oneEntity.ID) ([]oneEntity.ID, error) {
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

func (S SQLTeam) FindTeams(teamIDs []oneEntity.ID) ([]entity.Team, error) {
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
		err = rows.Scan(&team.ID, &team.Name, &team.LogoURL, &team.CreatedAt, &team.UpdatedAt)
		if err != nil {
			log.Println(team.ID, err)
			continue
		}
		teams = append(teams, team)
	}

	return teams, nil
}

func (S SQLTeam) AddUserToTeam(userID oneEntity.ID, teamID oneEntity.ID) (bool, error) {
	statement := `
	INSERT INTO team_member(
		team_id,
		user_id
	)
	VALUES ($1, $2);
`
	_, err := S.db.Exec(statement, teamID, userID)

	if err != nil {
		log.Println(err)
		return false, err
	}

	return true, nil
}

func (S SQLTeam) TeamExists(teamID oneEntity.ID) (bool, error) {
	query := fmt.Sprintf(`SELECT COUNT(1)  FROM team WHERE id = (%s)`, strconv.Itoa(int(teamID)))

	var count int
	err := S.db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, errors.WithStack(err)
	}

	if count == 0 {
		return false, nil
	}
	return true, nil
}

func (S SQLTeam) IsUserInTeam(userID oneEntity.ID, teamID oneEntity.ID) (bool, error) {
	query := fmt.Sprintf(`SELECT COUNT(1)  FROM team_member WHERE user_id = (%s) AND team_id = (%s) `, strconv.Itoa(int(userID)), strconv.Itoa(int(teamID)))

	var count int
	err := S.db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, errors.WithStack(err)
	}

	if count == 0 { // user is not in team
		return false, nil
	}

	return true, nil
}

func NewSQLTeam(db *sql.DB) SQLTeam {
	return SQLTeam{db: db}
}
