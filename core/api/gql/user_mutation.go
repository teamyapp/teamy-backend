package gql

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

func (m Mutation) CreateUser(ct context.Context, args struct {
	User struct {
		LastName   string
		FirstName  string
		ProfileURL *string
	}
}) (User, error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return User{}, err
	}

	user := entity.User{
		ID:         userID,
		CreatedAt:  time.Now(),
		FirstName:  args.User.FirstName,
		LastName:   args.User.LastName,
		ProfileURL: args.User.ProfileURL,
	}

	err := m.deps.userDao.CreateUser(ct, user)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return User{}, err
	}

	user, err := m.deps.userDao.FindUserByID(ct, userID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return User{}, err
	}

	user.FirstName = args.Input.FirstName
	user.LastName = args.Input.LastName
	updatedAt := time.Now()
	user.UpdatedAt = &updatedAt
	// TODO move this to user service
	realTimeTransaction := realtime.NewTransaction(m.deps.dataCollector, m.deps.stateSyncer)
	userMutation := mutation.NewUpdateUserMutation(
		m.deps.dataCollector,
		m.deps.stateSyncer,
		m.deps.teamMemberDao,
		m.deps.userDao,
		user)
	realTimeTransaction.AddMutation(ct, userMutation)
	err = realTimeTransaction.Commit(ct)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return User{}, err
	}

	return newUser(m.deps, user), nil
}

func (m Mutation) CreateUserProfileUploadSession(ct context.Context) (graphql.ID, error) {
	uploadSessionID, err := m.deps.userService.CreateUserProfileUploadSession(ct)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	return graphql.ID(strconv.FormatUint(uploadSessionID, 10)), nil
}

func (m Mutation) FinishUserProfileUploadSession(ct context.Context, args struct {
	FileUploadSessionID graphql.ID
}) (User, error) {
	fileUploadSessionID, err := fromGraphQLID(args.FileUploadSessionID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return User{}, err
	}

	user, err := m.deps.userService.FinishUserProfileUploadSession(ct, fileUploadSessionID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return User{}, err
	}

	return newUser(m.deps, user), nil
}
