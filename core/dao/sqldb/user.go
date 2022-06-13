package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type User struct {
	db *sql.DB
}

var _ dao.User = (*User)(nil)

func (u User) FindUserByID(userID uint64) (entity.User, error) {
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
	user := entity.User{}
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
		return entity.User{}, dao.ErrNotFound(fmt.Sprintf(
			"user not found: id=%v",
			userID))
	}

	return user, err
}

func (u User) FindUsersByIDs(userIDs []uint64) ([]entity.User, error) {
	if len(userIDs) == 0 {
		return []entity.User{}, nil
	}

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

	var users []entity.User
	for rows.Next() {
		var user entity.User
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

func (u User) CreateUser(user entity.User) error {
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

func (u User) UpdateUser(user entity.User) error {
	_, err := u.db.Exec(`
		UPDATE "user"
		SET
			first_name = $1,
			last_name = $2,
			profile_url = $3,
			updated_at = $4
		WHERE id = $5;`,
		user.FirstName,
		user.LastName,
		user.ProfileURL,
		user.UpdatedAt,
		user.ID,
	)
	return err
}

func NewUser(sqlDB *sql.DB) User {
	return User{db: sqlDB}
}
