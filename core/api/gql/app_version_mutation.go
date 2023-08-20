package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
)

func (m Mutation) CreateAppVersion(
	ctx context.Context,
	args struct {
		AppID graphql.ID
	}) AppVersion {
	panic("not implemented")
}

func (m Mutation) CreateAppPackageUploadSession(ct context.Context, args struct {
	AppID         graphql.ID
	VersionNumber int32
}) (graphql.ID, error) {
	appID, argErr := fromGraphQLID(args.AppID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.Error(internalErr)
		return "", errs.ToResolverErr(internalErr)
	}

	fileUploadSessionID, err := m.deps.appService.CreateAppPackageFileUploadSession(ct, appID, args.VersionNumber)
	if err != nil {
		m.deps.logger.Error(err)
		return "", errs.ToResolverErr(err)
	}

	return toGraphQLID(fileUploadSessionID), nil
}

func (m Mutation) FinishAppPackageUploadSession(ct context.Context, args struct {
	AppID               graphql.ID
	VersionNumber       int32
	FileUploadSessionID graphql.ID
}) (AppVersion, error) {
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

	appVersion, err := m.deps.appService.FinishAppPackageFileUploadSession(ct, appID, args.VersionNumber, fileUploadSessionID)

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
