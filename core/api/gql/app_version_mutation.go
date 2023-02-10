package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/service"
)

func (m Mutation) CreateAppVersion(ct context.Context, args struct {
	AppID graphql.ID
}) (AppVersion, error) {
	appID, argErr := fromGraphQLID(args.AppID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return AppVersion{}, errs.ToResolverErr(internalErr)
	}

	appVersion, err := m.deps.appService.CreateAppVersion(ct, appID)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return AppVersion{}, errs.ToResolverErr(err)
	}

	return newAppVersion(m.deps, appVersion), nil
}

func (m Mutation) UpdateAppVersion(ct context.Context, args struct {
	AppID         graphql.ID
	VersionNumber int32
	Input         struct {
		IconURL                   *string
		HasUIExtension            bool
		UIExtensionEntrypointPath *string
		Changes                   *string
		IsPublic                  bool
	}
}) (AppVersion, error) {
	appID, argErr := fromGraphQLID(args.AppID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return AppVersion{}, errs.ToResolverErr(internalErr)
	}

	input := service.UpdateAppVersionInput{
		IconURL:                   args.Input.IconURL,
		HasUIExtension:            args.Input.HasUIExtension,
		UIExtensionEntryPointPath: args.Input.UIExtensionEntrypointPath,
		Changes:                   args.Input.Changes,
		IsPublic:                  args.Input.IsPublic,
	}
	appVersion, err := m.deps.appService.UpdateAppVersion(ct, appID, args.VersionNumber, input)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return AppVersion{}, errs.ToResolverErr(err)
	}

	return newAppVersion(m.deps, appVersion), nil
}

func (m Mutation) DeleteAppVersion(ct context.Context, args struct {
	AppID         graphql.ID
	VersionNumber int32
}) (AppVersion, error) {
	appID, argErr := fromGraphQLID(args.AppID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return AppVersion{}, errs.ToResolverErr(internalErr)
	}

	appVersion, err := m.deps.appService.DeleteAppVersion(ct, appID, args.VersionNumber)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return AppVersion{}, errs.ToResolverErr(err)
	}

	return newAppVersion(m.deps, appVersion), nil
}

func (m Mutation) AddVisibleTeamToAppVersion(ct context.Context, args struct {
	AppID         graphql.ID
	VersionNumber int32
	TeamID        graphql.ID
}) (AppVersion, error) {
	appID, argErr := fromGraphQLID(args.AppID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return AppVersion{}, errs.ToResolverErr(internalErr)
	}

	teamID, argErr := fromGraphQLID(args.TeamID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return AppVersion{}, errs.ToResolverErr(internalErr)
	}

	appVersion, err := m.deps.appService.CreateAppVersionVisibleTeam(ct, appID, args.VersionNumber, teamID)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return AppVersion{}, errs.ToResolverErr(err)
	}

	return newAppVersion(m.deps, appVersion), nil
}

func (m Mutation) RemoveVisibleTeamFromAppVersion(ct context.Context, args struct {
	AppID         graphql.ID
	VersionNumber int32
	TeamID        graphql.ID
}) (AppVersion, error) {
	appID, argErr := fromGraphQLID(args.AppID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return AppVersion{}, errs.ToResolverErr(internalErr)
	}

	teamID, argErr := fromGraphQLID(args.TeamID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return AppVersion{}, errs.ToResolverErr(internalErr)
	}

	appVersion, err := m.deps.appService.DeleteAppVersionVisibleTeam(ct, appID, args.VersionNumber, teamID)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return AppVersion{}, errs.ToResolverErr(err)
	}

	return newAppVersion(m.deps, appVersion), nil
}
