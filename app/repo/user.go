package repo

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type User interface {
	FindUsers([]oneEntity.ID) ([]entity.User, error)
	FindUser(oneEntity.ID) (entity.User, error)
}

type SQLUser struct {
	db *sql.DB
}

var _ User = (*SQLUser)(nil)

func (S SQLUser) FindUsers(userIDs []oneEntity.ID) ([]entity.User, error) {
	idsString := toIDsString(userIDs)

	query := fmt.Sprintf(`
SELECT id, first_name, last_name, profile_url, created_at, updated_at
FROM "user"
WHERE id IN (%s)`, idsString)
	rows, err := S.db.Query(query)
	if err != nil {
		log.Println(err)
		return nil, err
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
		log.Println(user.ID, err)
		return entity.User{}, err
	}

	return user, nil
}

func NewSQLUser(db *sql.DB) SQLUser {
	return SQLUser{db: db}
}
