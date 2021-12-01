package repo

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
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
	idStrings := make([]string, 0)
	for _, singleID := range ids {
		idStrings = append(idStrings, strconv.Itoa(int(singleID)))
	}
	idsString := strings.Join(idStrings, ",")

	query := fmt.Sprintf(`SELECT * FROM "user" WHERE id IN (%s)`, idsString)

	rows, err := S.db.Query(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		err = rows.Scan(&user.ID, &user.FirstName, &user.LastName,&user.ProfileURL, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Println(user.ID, err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (S SQLUser) GetUser(userID oneEntity.ID) (entity.User, error) {
	query := fmt.Sprintf(`SELECT * FROM "user" WHERE id = (%s)`, strconv.Itoa(int(userID)))

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
