package repo

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/pkg/errors"
	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type User interface {
	FindUsers([]oneEntity.ID) ([]entity.User, error)
	FindUser(oneEntity.ID) (entity.User, error)
	UpdateActiveTeamId(userID oneEntity.ID, activeTeamID *oneEntity.ID) (*oneEntity.ID, error)
}

type SQLUser struct {
	db *sql.DB
}

var _ User = (*SQLUser)(nil)

func (S SQLUser) FindUsers(userIDs []oneEntity.ID) ([]entity.User, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}
	idsString := toIDsString(userIDs)

	query := fmt.Sprintf(`
SELECT id, first_name, last_name, profile_url, created_at, updated_at
FROM "user"
WHERE id IN (%s)`, idsString)
	rows, err := S.db.Query(query)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		err = rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.ProfileURL, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Println(user.ID, err)
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

func (S SQLUser) FindUser(userID oneEntity.ID) (entity.User, error) {
	query := fmt.Sprintf(`SELECT id, first_name, last_name, profile_url, created_at, updated_at FROM "user" WHERE id = (%s)`, strconv.Itoa(int(userID)))

	var user entity.User
	err := S.db.QueryRow(query).Scan(&user.ID, &user.FirstName, &user.LastName, &user.ProfileURL, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return entity.User{}, errors.WithStack(err)
	}

	return user, nil
}

func (S SQLUser) UpdateActiveTeamId(userID oneEntity.ID, activeTeamID *oneEntity.ID) (*oneEntity.ID, error) {
	statement := `
	UPDATE user_state AS updated
	SET active_team_id = $1
	FROM (
	    SELECT user_id, active_team_id
	    FROM user_state
	    WHERE user_id = $2
	    FOR UPDATE
	) AS previous
	WHERE updated.user_id = previous.user_id
	RETURNING previous.active_team_id;
`
	var previousActiveTeamID *int
	err := S.db.QueryRow(statement, activeTeamID, userID).Scan(&previousActiveTeamID)
	return (*oneEntity.ID)(previousActiveTeamID), errors.WithStack(err)
}

func NewSQLUser(db *sql.DB) SQLUser {
	return SQLUser{db: db}
}
