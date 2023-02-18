package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type App struct {
	deps *Dependencies
	app  entity.App
}

func (a App) ID() graphql.ID {
	return toGraphQLID(a.app.ID)
}

func (a App) APISecret() string {
	return a.app.APISecret
}

func (a App) ActiveVersion(ct context.Context) (*AppVersion, error) {
	if a.app.ActiveVersionNumber == nil {
		return nil, nil
	}

	appVersion, err := a.deps.appService.FindAppVersionByAppIDAndVersionNumber(ct, a.app.ID, *a.app.ActiveVersionNumber)
	if err != nil {
		a.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	activeVersion := newAppVersion(a.deps, appVersion)
	return &activeVersion, nil
}

func (a App) Name() string {
	return a.app.Name
}

func (a App) Versions(ct context.Context) ([]AppVersion, error) {
	appVersions, err := a.deps.appService.FindAppVersionsByAppID(ct, a.app.ID)
	if err != nil {
		a.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(appVersions, func(appVersion entity.AppVersion, _ int) AppVersion {
		return newAppVersion(a.deps, appVersion)
	}), nil
}

func (a App) TeamInstallations(ct context.Context) ([]AppTeamInstallation, error) {
	appTeamInstallations, err := a.deps.appService.FindAppTeamInstallationsByAppID(ct, a.app.ID)
	if err != nil {
		a.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(appTeamInstallations, func(appTeamInstallation entity.AppTeamInstallation, _ int) AppTeamInstallation {
		return newAppTeamInstallation(a.deps, appTeamInstallation)
	}), nil
}

func (a App) InstallationCount() int32 {
	return int32(a.app.InstallationCount)
}

func (a App) Description() string {
	return a.app.Description
}

func (a App) Creator(ct context.Context) (User, error) {
	user, err := a.deps.userService.FindUserByID(ct, a.app.CreatorUserID)
	if err != nil {
		a.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return User{}, errs.ToResolverErr(err)
	}

	return newUser(a.deps, user), nil
}

func (a App) CreatedAt() graphql.Time {
	return toGraphQLTime(a.app.CreatedAt)
}

func (a App) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(a.app.UpdatedAt)
}

func newApp(deps *Dependencies, app entity.App) App {
	return App{
		deps: deps,
		app:  app,
	}
}
