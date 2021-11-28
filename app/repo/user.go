package repo

import (
	"database/sql"
	"fmt"
	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
	"log"
	"strconv"
)

type User interface {
	GetUsers([]oneEntity.ID) ([]entity.User, error)
	GetUser(oneEntity.ID) (entity.User, error)
}

type SQLUser struct {
	db *sql.DB
}

var _ User = (*SQLUser)(nil)

func (S SQLUser) GetUsers(ids []oneEntity.ID) ([]entity.User, error) {
	var idsString string
	for _, singleID := range ids {
		idsString += strconv.Itoa(int(singleID)) + ","
	}
	idsString = idsString[:len(idsString) - 1]

	query := fmt.Sprintf("SELECT * FROM \"user\" WHERE id IN (%s)", idsString)

	rows, err := S.db.Query(query)


	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		err = rows.Scan(&user.ID, &user.Name, &user.ProfileURL, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Println(user.ID, err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (S SQLUser) GetUser(userID oneEntity.ID) (entity.User, error) {
	query := fmt.Sprintf("SELECT * FROM \"user\" WHERE id = (%s)", strconv.Itoa(int(userID)))

	var user entity.User
	err := S.db.QueryRow(query).Scan(&user.ID, &user.Name, &user.ProfileURL, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		log.Println(user.ID, err)
		return entity.User{}, err
	}

	return user, nil
}

func NewSQLUser(db *sql.DB) SQLUser {
	return SQLUser{db: db}
}
