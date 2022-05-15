package sqldb

import (
	"database/sql"

	"github.com/teamyapp/teamy-backend/app/dao"
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type User struct {
	db *sql.DB
}

var _ dao.User = (*User)(nil)

func (u User) FindUserByID(id uint64) (entityv2.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u User) FindUsersByIDs(ids []uint64) ([]entityv2.User, error) {
	//TODO implement me
	panic("implement me")
}

func NewUser(sqlDB *sql.DB) User {
	return User{db: sqlDB}
}
