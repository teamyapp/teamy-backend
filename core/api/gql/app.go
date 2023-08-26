package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type App struct {
	deps *Dependencies
	app  entity.App
}

func (a App) ID(ctx context.Context) graphql.ID {
	return toGraphQLID(a.app.ID)
}

func (a App) Secrets(ctx context.Context) []AppSecret {
	panic("not implemented")
}

func (a App) TotalInstallations(ctx context.Context) int32 {
	return int32(a.app.TotalInstallations)
}

func (a App) Installations(ctx context.Context) []TeamAppInstallation {
	panic("not implemented")
}

func (a App) Versions(ctx context.Context) []AppVersion {
	appVersions, err := a.deps.appService.FindAppVersionsByAppID(ctx, a.app.ID)
	if err != nil {
		a.deps.logger.ErrorWithContext(ctx, err)
		return []AppVersion{}
	}

	return collect.Map(appVersions, func(appVersion entity.AppVersion, index int) AppVersion {
		return newAppVersion(a.deps, appVersion)
	})
}

func (a App) UserGroups(ctx context.Context) []UserGroup {
	panic("not implemented")
}

func (a App) TeamGroups(ctx context.Context) []TeamGroup {
	panic("not implemented")
}

func (a App) UserRollouts(ctx context.Context) []UserRollout {
	panic("not implemented")
}

func (a App) TeamRollouts(ctx context.Context) []TeamRollout {
	panic("not implemented")
}

func (a App) ManagedByTeam(ctx context.Context) Team {
	teamID := a.app.ManagedByTeamID
	team, err := a.deps.teamService.FindTeamByID(ctx, teamID)
	if err != nil {
		a.deps.logger.ErrorWithContext(ctx, err)
		return newTeam(a.deps, entity.Team{})
	}

	return newTeam(a.deps, team)
}

func (m Mutation) CreateApp(
	ctx context.Context,
	args struct {
		TeamID graphql.ID
		Name   string
	}) App {
	teamID, internalErr := fromGraphQLID(args.TeamID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return App{}
	}

	app, err := m.deps.appService.CreateApp(ctx, args.Name, teamID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return App{}
	}

	return newApp(m.deps, app)
}

func (m Mutation) DeleteApp(
	ctx context.Context,
	args struct {
		AppID graphql.ID
	}) App {
	appID, internalErr := fromGraphQLID(args.AppID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return App{}
	}

	app, err := m.deps.appService.DeleteApp(ctx, appID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return App{}
	}

	return newApp(m.deps, app)
}

func (m Mutation) InstallAppToTeam(
	ctx context.Context,
	args struct {
		AppID  graphql.ID
		TeamID graphql.ID
	}) TeamAppInstallation {
	panic("not implemented")
}

func (m Mutation) UninstallAppFromTeam(
	ctx context.Context,
	args struct {
		InstallationID graphql.ID
	}) TeamAppInstallation {
	panic("not implemented")
}

func newApp(deps *Dependencies, app entity.App) App {
	return App{deps: deps, app: app}
}
