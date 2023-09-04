package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppSecret struct {
	deps      *Dependencies
	appSecret entity.AppSecret
}

func (a AppSecret) ID(ctx context.Context) graphql.ID {
	return toGraphQLID(a.appSecret.ID)
}

func (a AppSecret) Name(ctx context.Context) string {
	return a.appSecret.Name
}

func (a AppSecret) AddedAt(ctx context.Context) graphql.Time {
	return toGraphQLTime(a.appSecret.AddedAt)
}

func (a AppSecret) AddedBy(ctx context.Context) (User, error) {
	user, err := a.deps.userService.FindUserByID(ctx, a.appSecret.AddedByUserID)
	if err != nil {
		a.deps.logger.ErrorWithContext(ctx, err)
		return User{}, errs.ToResolverErr(err)
	}

	return newUser(a.deps, user), nil
}

func (a AppSecret) LastUsedAt(ctx context.Context) *graphql.Time {
	return toGraphQLTimePtr(a.appSecret.LastUsedAt)
}

func (a AppSecret) App(ctx context.Context) (App, error) {
	app, err := a.deps.appService.FindAppByID(ctx, a.appSecret.AppID)
	if err != nil {
		a.deps.logger.ErrorWithContext(ctx, err)
		return App{}, errs.ToResolverErr(err)
	}

	return newApp(a.deps, app), nil
}

func (m Mutation) CreateAppSecret(
	ctx context.Context,
	args struct {
		AppID graphql.ID
		Input struct {
			Name string
		}
	}) (AppSecret, error) {
	appID, argErr := fromGraphQLID(args.AppID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.Error(internalErr)
		return AppSecret{}, errs.ToResolverErr(internalErr)
	}

	appSecret, err := m.deps.appService.CreateAppSecret(ctx, appID, args.Input.Name)
	if err != nil {
		m.deps.logger.Error(err)
		return AppSecret{}, errs.ToResolverErr(err)
	}

	return newAppSecret(m.deps, appSecret), nil
}

func (m Mutation) UpdateAppSecret(
	ctx context.Context,
	args struct {
		SecretID graphql.ID
		Input    struct {
			Name string
		}
	}) (AppSecret, error) {
	appID, internalErr := fromGraphQLID(args.SecretID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return AppSecret{}, errs.ToResolverErr(internalErr)
	}

	secretID, internalErr := fromGraphQLID(args.SecretID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return AppSecret{}, errs.ToResolverErr(internalErr)
	}

	appSecret, err := m.deps.appService.UpdateAppSecret(ctx, appID, secretID, args.Input.Name)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return AppSecret{}, errs.ToResolverErr(err)
	}

	return newAppSecret(m.deps, appSecret), nil
}

func (m Mutation) DeleteAppSecret(
	ctx context.Context,
	args struct {
		SecretID graphql.ID
	},
) (AppSecret, error) {
	secretID, internalErr := fromGraphQLID(args.SecretID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return AppSecret{}, errs.ToResolverErr(internalErr)
	}

	appSecret, err := m.deps.appService.DeleteAppSecret(ctx, secretID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return AppSecret{}, errs.ToResolverErr(err)
	}

	return newAppSecret(m.deps, appSecret), nil
}

func newAppSecret(deps *Dependencies, appSecret entity.AppSecret) AppSecret {
	return AppSecret{deps: deps, appSecret: appSecret}
}
