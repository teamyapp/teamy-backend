package gql

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/teamy-backend/core/entity"
)

func (m Mutation) CreateUser(ct context.Context, args struct {
	User struct {
		LastName   string
		FirstName  string
		ProfileURL *string
	}
}) (User, error) {
	userID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		return User{}, err
	}

	user := entity.User{
		ID:         userID,
		CreatedAt:  time.Now(),
		FirstName:  args.User.FirstName,
		LastName:   args.User.LastName,
		ProfileURL: args.User.ProfileURL,
	}

	err = m.deps.userDao.CreateUser(user)
	if err != nil {
		return User{}, err
	}

	return newUser(m.deps, user), nil
}

func (m Mutation) UpdateUser(ct context.Context, args struct {
	UserID graphql.ID
	Input  struct {
		LastName  string
		FirstName string
	}
}) (User, error) {
	userID, err := fromGraphQLID(args.UserID)
	if err != nil {
		return User{}, err
	}

	user, err := m.deps.userDao.FindUserByID(userID)
	if err != nil {
		return User{}, err
	}

	user.FirstName = args.Input.FirstName
	user.LastName = args.Input.LastName
	updatedAt := time.Now()
	user.UpdatedAt = &updatedAt
	err = m.deps.userSyncer.UpdateAndSyncUser(user)
	if err != nil {
		return User{}, err
	}

	return newUser(m.deps, user), nil
}

func (m Mutation) CreateUserProfileUploadSession(ct context.Context, args struct {
	UserID graphql.ID
}) (graphql.ID, error) {
	// TODO: implement me
	panic("implement me")
}

func (m Mutation) FinishUserProfileUploadSession(ct context.Context, args struct {
	UserID graphql.ID
}) (FileMetadata, error) {
	// TODO: implement me
	panic("implement me")
}
