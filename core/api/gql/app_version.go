package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppVersion struct {
	deps       *Dependencies
	appVersion entity.AppVersion
}

func (a AppVersion) Number(ctx context.Context) int32 {
	return int32(a.appVersion.Number)
}

func (a AppVersion) AppName(ctx context.Context) string {
	panic("not implemented")
}

func (a AppVersion) Description(ctx context.Context) string {
	panic("not implemented")
}

func (a AppVersion) Changes(ctx context.Context) []string {
	panic("not implemented")
}

func (a AppVersion) CreatedAt(ctx context.Context) graphql.Time {
	panic("not implemented")
}

func (a AppVersion) CreatedBy(ctx context.Context) User {
	panic("not implemented")
}

func (a AppVersion) Prices(ctx context.Context) []Money {
	panic("not implemented")
}

func (a AppVersion) IsReady(ctx context.Context) bool {
	panic("not implemented")
}

func (a AppVersion) App(ctx context.Context) App {
	panic("not implemented")
}

func newAppVersion(deps *Dependencies, appVersion entity.AppVersion) AppVersion {
	return AppVersion{deps: deps, appVersion: appVersion}
}

func (m Mutation) CreateAppVersion(
	ctx context.Context,
	args struct {
		AppID graphql.ID
	}) AppVersion {
	panic("not implemented")
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
		m.deps.logger.Error(internalErr)
		return "", errs.ToResolverErr(internalErr)
	}

	fileUploadSessionID, err := m.deps.appService.CreateAppPackageFileUploadSession(ct, appID, int(args.VersionNumber))
	if err != nil {
		m.deps.logger.Error(err)
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
		return AppVersion{}, errs.ToResolverErr(internalErr)
	}

	fileUploadSessionID, argErr := fromGraphQLID(args.FileUploadSessionID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		return AppVersion{}, errs.ToResolverErr(internalErr)
	}

	appVersion, err := m.deps.appService.FinishAppPackageFileUploadSession(ct, appID, int(args.VersionNumber), fileUploadSessionID)
	if err != nil {
		m.deps.logger.Error(err)
		return AppVersion{}, errs.ToResolverErr(err)
	}

	return newAppVersion(m.deps, appVersion), nil
}

func (m Mutation) DeleteAppVersion(
	ctx context.Context,
	args struct {
		AppID         graphql.ID
		VersionNumber int32
	}) AppVersion {
	panic("not implemented")
}
