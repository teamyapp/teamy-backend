package datastore

import (
	"fmt"

	"github.com/pkg/errors"
	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

//
// User
//
func (d DataStore) GetUser(id oneEntity.ID) (entity.User, error) {
	user, ok := d.data.Users[id]
	if !ok {
		return entity.User{}, fmt.Errorf("user %v is not found", id)
	}
	return user, nil
}

func (d DataStore) GetUsers(ids []oneEntity.ID) (users []entity.User, err error) {
	for _, id := range ids {
		user, ok := d.data.Users[id]
		if ok {
			if user.ID == id {
				users = append(users, user)
			} else {
				err = errors.Errorf("user key %v and id %v doesn't match", id, user.ID)
				return
			}
		}
	}
	return
}

func (d DataStore) CreateUser(id oneEntity.ID) (entity.User, error) {
	if _, ok := d.data.Users[id]; ok {
		return entity.User{}, errors.Errorf("user %v already exists", id)
	}
	d.data.Users[id] = entity.User{
		Entity: oneEntity.Entity{
			ID: id,
		},
	}
	return d.data.Users[id], d.persister.Write(d.data)
}

func (d DataStore) UpdateUser(userID oneEntity.ID, apply func(entity.User) entity.User) (entity.User, error) {
	_, ok := d.data.Users[userID]
	if !ok {
		return entity.User{}, errors.Errorf("user %v is not found", userID)
	}
	d.data.Users[userID] = apply(d.data.Users[userID])
	return d.data.Users[userID], nil
}
