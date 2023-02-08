package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppVersion struct {
	deps       *Dependencies
	appVersion entity.AppVersion
}

func (a AppVersion) App(ct context.Context) (App, error) {
	panic("implement me")
}

func (a AppVersion) VersionNumber() int32 {
	return a.appVersion.VersionNumber
}

func (a AppVersion) IconURL() *string {
	return a.appVersion.IconURL
}

func (a AppVersion) HasUIExtension() bool {
	return a.appVersion.HasUIExtension
}

func (a AppVersion) UIExtensionEntrypointPath() *string {
	return a.appVersion.UIExtensionEntrypointPath
}

func (a AppVersion) IsPublic() bool {
	return a.appVersion.IsPublic
}

func (a AppVersion) VisibleToTeams(ct context.Context) []Team {
	appVersionVisibleTeams, err := a.deps.appService.FindAppVersionVisibleTeams(ct, a.appVersion.AppID, a.appVersion.VersionNumber)
	if err != nil {
		a.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil
	}

	teamIDs := collect.Map(appVersionVisibleTeams, func(appVersionVisibleTeam entity.AppVersionVisibleTeam, _ int) uint64 {
		return appVersionVisibleTeam.TeamID
	})

	var teams []entity.Team
	for _, teamID := range teamIDs {
		team, err := a.deps.teamService.FindTeamByID(ct, teamID)
		if err != nil {
			a.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			continue
		}

		teams = append(teams, team)
	}

	return collect.Map(teams, func(team entity.Team, _ int) Team {
		return newTeam(a.deps, team)
	})
}

func (a AppVersion) Changes() *string {
	return a.appVersion.Changes
}

func (a AppVersion) CreatedAt() graphql.Time {
	return toGraphQLTime(a.appVersion.CreatedAt)
}

func (a AppVersion) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(a.appVersion.UpdateAt)
}

func newAppVersion(deps *Dependencies, appVersion entity.AppVersion) AppVersion {
	return AppVersion{
		deps:       deps,
		appVersion: appVersion,
	}
}
