package gql

import (
	"context"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/service"
)

func (m Mutation) CreateUser(ct context.Context, args struct {
	User struct {
		LastName   string
		FirstName  string
		ProfileURL *string
	}
}) (User, error) {
	input := service.CreateUserInput{
		LastName:   args.User.LastName,
		FirstName:  args.User.FirstName,
		ProfileURL: args.User.ProfileURL,
	}
	user, err := m.deps.userService.CreateUser(ct, input)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
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
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return User{}, err
	}

	input := service.UpdateUserInput{
		LastName:  args.Input.LastName,
		FirstName: args.Input.FirstName,
	}
	user, err := m.deps.userService.UpdateUser(ct, userID, input)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return User{}, err
	}

	return newUser(m.deps, user), nil
}

func (m Mutation) CreateUserProfileUploadSession(ct context.Context) (graphql.ID, error) {
	uploadSessionID, err := m.deps.userService.CreateUserProfileUploadSession(ct)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return "", err
	}

	return graphql.ID(strconv.FormatUint(uploadSessionID, 10)), nil
}

func (m Mutation) FinishUserProfileUploadSession(ct context.Context, args struct {
	FileUploadSessionID graphql.ID
}) (User, error) {
	fileUploadSessionID, err := fromGraphQLID(args.FileUploadSessionID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return User{}, err
	}

	user, err := m.deps.userService.FinishUserProfileUploadSession(ct, fileUploadSessionID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return User{}, err
	}

	return newUser(m.deps, user), nil
}
