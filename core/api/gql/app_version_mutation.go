package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/service"
)

func (m Mutation) CreateAppVersion(ct context.Context, args struct {
	AppID graphql.ID
}) (AppVersion, error) {
	appID, argErr := fromGraphQLID(args.AppID)
	if argErr != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: argErr})
		return AppVersion{}, argErr
	}

	appVersion, err := m.deps.appService.CreateAppVersion(ct, appID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
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
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: argErr})
		return AppVersion{}, argErr
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
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
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
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: argErr})
		return AppVersion{}, argErr
	}

	appVersion, err := m.deps.appService.DeleteAppVersion(ct, appID, args.VersionNumber)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
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
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: argErr})
		return AppVersion{}, argErr
	}

	teamID, argErr := fromGraphQLID(args.TeamID)
	if argErr != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: argErr})
		return AppVersion{}, argErr
	}

	appVersionVisibleTeam, err := m.deps.appService.CreateAppVersionVisibleTeam(ct, appID, args.VersionNumber, teamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return AppVersion{}, errs.ToResolverErr(err)
	}

	appVersion, err := m.deps.appService.FindAppVersionByAppIDAndVersionNumber(ct, appVersionVisibleTeam.AppID, appVersionVisibleTeam.VersionNumber)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
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
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: argErr})
		return AppVersion{}, argErr
	}

	teamID, argErr := fromGraphQLID(args.TeamID)
	if argErr != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: argErr})
		return AppVersion{}, argErr
	}

	appVersionVisibleTeam, err := m.deps.appService.DeleteAppVersionVisibleTeam(ct, appID, args.VersionNumber, teamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return AppVersion{}, errs.ToResolverErr(err)
	}

	appVersion, err := m.deps.appService.FindAppVersionByAppIDAndVersionNumber(ct, appVersionVisibleTeam.AppID, appVersionVisibleTeam.VersionNumber)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return AppVersion{}, errs.ToResolverErr(err)
	}

	return newAppVersion(m.deps, appVersion), nil
}
