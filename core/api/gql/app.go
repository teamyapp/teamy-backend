package gql

import (
	"context"
	"fmt"

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

func (a App) Secrets(ctx context.Context) ([]AppSecret, error) {
	secrets, err := a.deps.appService.FindSecretsByAppID(ctx, a.app.ID)
	if err != nil {
		a.deps.logger.ErrorWithContext(ctx, err)
		return []AppSecret{}, errs.ToResolverErr(err)
	}

	return collect.Map(secrets, func(appSecret entity.AppSecret, index int) AppSecret {
		return newAppSecret(a.deps, appSecret)
	}), nil
}

func (a App) TotalInstallations(ctx context.Context) int32 {
	return int32(a.app.TotalInstallations)
}

func (a App) Installations(ctx context.Context) ([]TeamAppInstallation, error) {
	teamAppInstallations, err := a.deps.appService.FindTeamAppInstallationsByAppID(ctx, a.app.ID)
	if err != nil {
		a.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(teamAppInstallations, func(teamAppInstallation entity.TeamAppInstallation, index int) TeamAppInstallation {
		return newTeamAppInstallation(a.deps, teamAppInstallation)
	}), nil
}

func (a App) Versions(ctx context.Context) ([]AppVersion, error) {
	appVersions, err := a.deps.appService.FindAppVersionsByAppID(ctx, a.app.ID)
	if err != nil {
		a.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(appVersions, func(appVersion entity.AppVersion, index int) AppVersion {
		return newAppVersion(a.deps, appVersion)
	}), nil
}

func (a App) UserGroups(ctx context.Context) ([]Group, error) {
	groups, err := a.deps.groupService.FindUserGroupsByAppID(ctx, a.app.ID)
	if err != nil {
		a.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	userGroups := make([]Group, 0)
	for _, group := range groups {
		switch group.Type {
		case entity.GroupTypeStatic:
			userGroups = append(userGroups, newStaticUserGroup(a.deps, group.StaticGroup))
		case entity.GroupTypeFilter:
			userGroups = append(userGroups, newFilterGroup(a.deps, group.FilterGroup))
		default:
			return nil, errs.ToResolverErr(errs.NewError(errs.Unknown, fmt.Sprintf("unknown group type %s", group.Type)))
		}
	}

	return userGroups, nil
}

func (a App) TeamGroups(ctx context.Context) ([]Group, error) {
	groups, err := a.deps.groupService.FindTeamGroupsByAppID(ctx, a.app.ID)
	if err != nil {
		a.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	teamGroups := make([]Group, 0)
	for _, group := range groups {
		switch group.Type {
		case entity.GroupTypeStatic:
			teamGroups = append(teamGroups, newStaticTeamGroup(a.deps, group.StaticGroup))
		case entity.GroupTypeFilter:
			teamGroups = append(teamGroups, newFilterGroup(a.deps, group.FilterGroup))
		default:
			return nil, errs.ToResolverErr(errs.NewError(errs.Unknown, fmt.Sprintf("unknown group type %s", group.Type)))
		}
	}

	return teamGroups, nil
}

func (a App) UserRollouts(ctx context.Context) ([]Rollout, error) {
	userRollouts, err := a.deps.rolloutService.FindUserRolloutsByAppID(ctx, a.app.ID)
	if err != nil {
		a.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(userRollouts, func(userRollout entity.Rollout, index int) Rollout {
		return newRollout(a.deps, userRollout)
	}), nil
}

func (a App) TeamRollouts(ctx context.Context) ([]Rollout, error) {
	teamRollouts, err := a.deps.rolloutService.FindTeamRolloutsByAppID(ctx, a.app.ID)
	if err != nil {
		a.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(teamRollouts, func(teamRollout entity.Rollout, index int) Rollout {
		return newRollout(a.deps, teamRollout)
	}), nil
}

func (a App) ManagedByTeam(ctx context.Context) (Team, error) {
	teamID := a.app.ManagedByTeamID
	team, err := a.deps.teamService.FindTeamByID(ctx, teamID)
	if err != nil {
		a.deps.logger.ErrorWithContext(ctx, err)
		return Team{}, errs.ToResolverErr(err)
	}

	return newTeam(a.deps, team), nil
}

func (a App) LatestVersionForTeam(
	ctx context.Context,
	args struct {
		TeamID graphql.ID
	}) (*AppVersion, error) {
	teamID, internalErr := fromGraphQLID(args.TeamID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		a.deps.logger.ErrorWithContext(ctx, internalErr)
		return nil, errs.ToResolverErr(internalErr)
	}

	appVersionNumber, err := a.deps.rolloutService.GetActiveAppVersionNumberForTeam(ctx, a.app.ID, teamID)
	if err != nil {
		return nil, errs.ToResolverErr(err)
	}

	if appVersionNumber == nil {
		return nil, nil
	}

	appVersion, err := a.deps.appService.FindAppVersionByAppIDAndNumber(ctx, a.app.ID, *appVersionNumber)
	if err != nil {
		return nil, errs.ToResolverErr(err)
	}

	gqlAppVersion := newAppVersion(a.deps, appVersion)
	return &gqlAppVersion, nil
}

func (a App) Tags(ctx context.Context) ([]Tag, error) {
	tags, err := a.deps.appService.FindTagsByAppID(ctx, a.app.ID)
	if err != nil {
		a.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(tags, func(tag entity.Tag, index int) Tag {
		return newTag(a.deps, tag)
	}), nil
}

func (m Mutation) CreateApp(
	ctx context.Context,
	args struct {
		TeamID graphql.ID
		Name   string
	}) (App, error) {
	teamID, internalErr := fromGraphQLID(args.TeamID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return App{}, errs.ToResolverErr(internalErr)
	}

	app, err := m.deps.appService.CreateApp(ctx, args.Name, teamID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return App{}, errs.ToResolverErr(err)
	}

	return newApp(m.deps, app), nil
}

func (m Mutation) DeleteApp(
	ctx context.Context,
	args struct {
		AppID graphql.ID
	}) (App, error) {
	appID, internalErr := fromGraphQLID(args.AppID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return App{}, errs.ToResolverErr(internalErr)
	}

	app, err := m.deps.appService.DeleteApp(ctx, appID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return App{}, errs.ToResolverErr(err)
	}

	return newApp(m.deps, app), nil
}

func (m Mutation) InstallAppToTeam(
	ctx context.Context,
	args struct {
		AppID  graphql.ID
		TeamID graphql.ID
	},
) (TeamAppInstallation, error) {
	appID, internalErr := fromGraphQLID(args.AppID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return TeamAppInstallation{}, errs.ToResolverErr(internalErr)
	}

	teamID, internalErr := fromGraphQLID(args.TeamID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return TeamAppInstallation{}, errs.ToResolverErr(internalErr)
	}

	teamAppInstallation, err := m.deps.appService.InstallAppToTeam(ctx, appID, teamID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return TeamAppInstallation{}, errs.ToResolverErr(err)
	}

	return newTeamAppInstallation(m.deps, teamAppInstallation), nil
}

func (m Mutation) UninstallAppFromTeam(
	ctx context.Context,
	args struct {
		InstallationID graphql.ID
	},
) (TeamAppInstallation, error) {
	teamAppInstallationID, internalErr := fromGraphQLID(args.InstallationID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return TeamAppInstallation{}, errs.ToResolverErr(internalErr)
	}

	teamAppInstallation, err := m.deps.appService.UninstallAppFromTeam(ctx, teamAppInstallationID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return TeamAppInstallation{}, errs.ToResolverErr(err)
	}

	return newTeamAppInstallation(m.deps, teamAppInstallation), nil
}

func newApp(deps *Dependencies, app entity.App) App {
	return App{deps: deps, app: app}
}
