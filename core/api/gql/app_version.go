package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/service"
)

type AppVersion struct {
	deps       *Dependencies
	appVersion entity.AppVersion
}

func (a AppVersion) Number(ctx context.Context) int32 {
	return int32(a.appVersion.Number)
}

func (a AppVersion) AppName(ctx context.Context) string {
	return a.appVersion.AppName
}

func (a AppVersion) Description(ctx context.Context) string {
	return a.appVersion.Description
}

func (a AppVersion) Changes(ctx context.Context) ([]string, error) {
	changes, err := a.deps.appService.FindAppVersionChangesByAppVersionID(ctx, a.appVersion.AppID, a.appVersion.Number)
	if err != nil {
		a.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return changes, nil
}

func (a AppVersion) CreatedAt(ctx context.Context) graphql.Time {
	return toGraphQLTime(a.appVersion.CreatedAt)
}

func (a AppVersion) CreatedBy(ctx context.Context) (User, error) {
	user, err := a.deps.userService.FindUserByID(ctx, a.appVersion.CreatedByUserID)
	if err != nil {
		a.deps.logger.ErrorWithContext(ctx, err)
		return User{}, errs.ToResolverErr(err)
	}

	return newUser(a.deps, user), nil
}

func (a AppVersion) Prices(ctx context.Context) ([]Money, error) {
	prices, err := a.deps.appService.FindPricesByAppVersionID(ctx, a.appVersion.AppID, a.appVersion.Number)
	if err != nil {
		a.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(prices, func(price entity.Money, index int) Money {
		return newMoney(a.deps, price)
	}), nil
}

func (a AppVersion) Status(ctx context.Context) entity.AppVersionStatus {
	return a.appVersion.Status
}

func (a AppVersion) Locked(ctx context.Context) bool {
	return a.appVersion.Locked
}

func (a AppVersion) App(ctx context.Context) (App, error) {
	app, err := a.deps.appService.FindAppByID(ctx, a.appVersion.AppID)
	if err != nil {
		a.deps.logger.ErrorWithContext(ctx, err)
		return App{}, errs.ToResolverErr(err)
	}

	return newApp(a.deps, app), nil
}

func (a AppVersion) ErrorMessage(ctx context.Context) *string {
	return a.appVersion.ErrorMessage
}

func newAppVersion(deps *Dependencies, appVersion entity.AppVersion) AppVersion {
	return AppVersion{deps: deps, appVersion: appVersion}
}

func (m Mutation) CreateAppVersion(
	ctx context.Context,
	args struct {
		AppID graphql.ID
	},
) (AppVersion, error) {
	appID, argErr := fromGraphQLID(args.AppID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return AppVersion{}, errs.ToResolverErr(internalErr)
	}

	appVersion, err := m.deps.appService.CreateAppVersion(ctx, appID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return AppVersion{}, errs.ToResolverErr(err)
	}

	return newAppVersion(m.deps, appVersion), nil
}

func (m Mutation) UpdateAppVersion(
	ctx context.Context,
	args struct {
		AppID         graphql.ID
		VersionNumber int32
		Input         struct {
			Status entity.AppVersionStatus
		}
	},
) (AppVersion, error) {
	appID, argErr := fromGraphQLID(args.AppID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return AppVersion{}, errs.ToResolverErr(internalErr)
	}

	updateAppVersionInput := service.UpdateAppVersionInput{
		Status: args.Input.Status,
	}

	appVersion, err := m.deps.appService.UpdateAppVersion(ctx, appID, int(args.VersionNumber), updateAppVersionInput)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return AppVersion{}, errs.ToResolverErr(err)
	}

	return newAppVersion(m.deps, appVersion), nil
}

func (m Mutation) CreateAppPackageUploadSession(
	ct context.Context,
	args struct {
		AppID         graphql.ID
		VersionNumber int32
	},
) (graphql.ID, error) {
	appID, argErr := fromGraphQLID(args.AppID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return "", errs.ToResolverErr(internalErr)
	}

	fileUploadSessionID, err := m.deps.appService.CreateAppPackageFileUploadSession(ct, appID, int(args.VersionNumber))
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return "", errs.ToResolverErr(err)
	}

	return toGraphQLID(fileUploadSessionID), nil
}

func (m Mutation) FinishAppPackageUploadSession(
	ct context.Context,
	args struct {
		AppID               graphql.ID
		VersionNumber       int32
		FileUploadSessionID graphql.ID
	},
) (AppVersion, error) {
	appID, argErr := fromGraphQLID(args.AppID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return AppVersion{}, errs.ToResolverErr(internalErr)
	}

	fileUploadSessionID, argErr := fromGraphQLID(args.FileUploadSessionID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return AppVersion{}, errs.ToResolverErr(internalErr)
	}

	appVersion, err := m.deps.appService.FinishAppPackageFileUploadSession(ct, appID, int(args.VersionNumber), fileUploadSessionID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return AppVersion{}, errs.ToResolverErr(err)
	}

	return newAppVersion(m.deps, appVersion), nil
}

func (m Mutation) DeleteAppVersion(
	ctx context.Context,
	args struct {
		AppID         graphql.ID
		VersionNumber int32
	},
) (AppVersion, error) {
	appID, argErr := fromGraphQLID(args.AppID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return AppVersion{}, errs.ToResolverErr(internalErr)
	}

	appVersion, err := m.deps.appService.DeleteAppVersion(ctx, appID, int(args.VersionNumber))
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return AppVersion{}, errs.ToResolverErr(err)
	}

	return newAppVersion(m.deps, appVersion), nil
}
