package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/teamyapp/teamy-backend/app/dao"
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type User struct {
	db *sql.DB
}

var _ dao.User = (*User)(nil)

func (u User) FindUserByID(userID uint64) (entityv2.User, error) {
	statement := `
	SELECT
		id,
		first_name,
		last_name,
		profile_url,
		created_at,
		updated_at
	FROM "user"
	WHERE id = $1;
`
	user := entityv2.User{}
	err := u.db.QueryRow(statement, userID).
		Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.ProfileURL,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entityv2.User{}, dao.ErrNotFound(fmt.Sprintf(
			"user not found: id=%v",
			userID))
	}

	return user, err
}

func (u User) FindUsersByIDs(userIDs []uint64) ([]entityv2.User, error) {
	idsString := toIDsString(userIDs)
	query := fmt.Sprintf(`
	SELECT
		id,
		first_name,
		last_name,
		profile_url,
		created_at,
		updated_at
	FROM "user"
	WHERE id IN (%s)`, idsString)
	rows, err := u.db.Query(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	var users []entityv2.User
	for rows.Next() {
		var user entityv2.User
		err = rows.
			Scan(
				&user.ID,
				&user.FirstName,
				&user.LastName,
				&user.ProfileURL,
				&user.CreatedAt,
				&user.UpdatedAt,
			)
		if err != nil {
			log.Println(user.ID, err)
			continue
		}

		users = append(users, user)
	}

	return users, nil
}

func (u User) CreateUser(user entityv2.User) error {
	_, err := u.db.Exec(`
		INSERT INTO "user"
		(
			id,
			first_name,
			last_name,
			profile_url,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5);`,
		user.ID,
		user.FirstName,
		user.LastName,
		user.ProfileURL,
		user.CreatedAt,
	)
	return err
}

func NewUser(sqlDB *sql.DB) User {
	return User{db: sqlDB}
}
