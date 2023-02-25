package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/service"
)

func (m Mutation) CreateApp(ct context.Context, args struct {
	Name string
}) (App, error) {
	app, err := m.deps.appService.CreateApp(ct, args.Name)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return App{}, errs.ToResolverErr(err)
	}

	return newApp(m.deps, app), nil
}

func (m Mutation) UpdateApp(ct context.Context, args struct {
	AppID graphql.ID
	Input struct {
		Name                *string
		ActiveVersionNumber *int32
		Description         *string
	}
}) (App, error) {
	appID, argErr := fromGraphQLID(args.AppID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return App{}, errs.ToResolverErr(internalErr)
	}

	input := service.UpdateAppInput{
		Name:                args.Input.Name,
		Description:         args.Input.Description,
		ActiveVersionNumber: args.Input.ActiveVersionNumber,
	}
	app, err := m.deps.appService.UpdateApp(ct, appID, input)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return App{}, errs.ToResolverErr(err)
	}

	return newApp(m.deps, app), nil
}

func (m Mutation) RefreshAppSecret(ct context.Context, args struct {
	AppID graphql.ID
}) (App, error) {
	appID, argErr := fromGraphQLID(args.AppID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return App{}, errs.ToResolverErr(internalErr)
	}

	app, err := m.deps.appService.RefreshAppSecret(ct, appID)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return App{}, errs.ToResolverErr(err)
	}

	return newApp(m.deps, app), nil
}

func (m Mutation) DeleteApp(ct context.Context, args struct {
	AppID graphql.ID
}) (App, error) {
	appID, argErr := fromGraphQLID(args.AppID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return App{}, errs.ToResolverErr(internalErr)
	}

	app, err := m.deps.appService.DeleteApp(ct, appID)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return App{}, errs.ToResolverErr(err)
	}

	return newApp(m.deps, app), nil
}
