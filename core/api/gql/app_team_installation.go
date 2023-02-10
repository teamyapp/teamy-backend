package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppTeamInstallation struct {
	deps                *Dependencies
	appTeamInstallation entity.AppTeamInstallation
}

func (a AppTeamInstallation) App() (App, error) {
	panic("implement me")
}

func (a AppTeamInstallation) EnabledVersion(ct context.Context) (AppVersion, error) {
	appVersion, err := a.deps.appService.FindAppVersionByAppIDAndVersionNumber(ct, a.appTeamInstallation.AppID, a.appTeamInstallation.EnabledVersionNumber)
	if err != nil {
		a.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return AppVersion{}, errs.ToResolverErr(err)
	}

	return newAppVersion(a.deps, appVersion), nil
}

func (a AppTeamInstallation) InstalledTeam(ct context.Context) (Team, error) {
	team, err := a.deps.teamService.FindTeamByID(ct, a.appTeamInstallation.InstalledTeamID)
	if err != nil {
		a.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return Team{}, errs.ToResolverErr(err)
	}

	return newTeam(a.deps, team), nil
}

func (a AppTeamInstallation) InstalledBy(ct context.Context) (*User, error) {
	if a.appTeamInstallation.InstalledByUserID == nil {
		return nil, nil
	}

	user, err := a.deps.userService.FindUserByID(ct, *a.appTeamInstallation.InstalledByUserID)
	if err != nil {
		a.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, errs.ToResolverErr(err)
	}

	appUser := newUser(a.deps, user)
	return &appUser, nil
}

func (a AppTeamInstallation) InstalledAt() graphql.Time {
	return toGraphQLTime(a.appTeamInstallation.InstalledAt)
}

func newAppTeamInstallation(deps *Dependencies, appTeamInstallation entity.AppTeamInstallation) AppTeamInstallation {
	return AppTeamInstallation{
		deps:                deps,
		appTeamInstallation: appTeamInstallation,
	}
}
