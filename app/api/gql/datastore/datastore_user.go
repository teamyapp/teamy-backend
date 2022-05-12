package datastore

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/teamyapp/teamy-backend/app/entity"
)

//
// User
//
func (d DataStore) GetUser(id uint64) (entity.User, error) {
	user, ok := d.data.Users[id]
	if !ok {
		return entity.User{}, fmt.Errorf("user %v is not found", id)
	}
	return user, nil
}

func (d DataStore) GetUsers(ids []uint64) (users []entity.User, err error) {
	for _, id := range ids {
		user, ok := d.data.Users[id]
		if ok {
			if user.ID == id {
				users = append(users, user)
			} else {
				err = errors.Errorf("user key %v and id %v doesn't match", id, user.ID)
				return
			}
		} else {
			users = append(users, entity.GhostUser())
		}
	}
	return
}

func (d DataStore) CreateUser(user entity.User) (entity.User, error) {
	if _, ok := d.data.Users[user.ID]; ok {
		return entity.User{}, errors.Errorf("user %v already exists", user.ID)
	}
	d.data.Users[user.ID] = user
	return user, d.persister.Write(d.data)
}

func (d DataStore) UpdateUser(userID uint64, apply func(entity.User) entity.User) (entity.User, error) {
	_, ok := d.data.Users[userID]
	if !ok {
		return entity.User{}, errors.Errorf("user %v is not found", userID)
	}
	d.data.Users[userID] = apply(d.data.Users[userID])
	return d.data.Users[userID], d.persister.Write(d.data)
}
