package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamAppInstallation struct {
	deps                *Dependencies
	teamAppInstallation entity.TeamAppInstallation
}

func (t TeamAppInstallation) ID() graphql.ID {
	return toGraphQLID(t.teamAppInstallation.ID)
}

func (t TeamAppInstallation) InstalledTeam(ctx context.Context) (Team, error) {
	team, err := t.deps.teamService.FindTeamByID(ctx, t.teamAppInstallation.InstalledTeamID)
	if err != nil {
		t.deps.logger.ErrorWithContext(ctx, err)
		return Team{}, errs.ToResolverErr(err)
	}

	return newTeam(t.deps, team), nil
}

func (t TeamAppInstallation) ActiveAppVersion(ctx context.Context) (*AppVersion, error) {
	appVersionNumber, err := t.deps.rolloutService.GetActiveAppVersionNumberForTeam(
		ctx,
		t.teamAppInstallation.AppID,
		t.teamAppInstallation.InstalledTeamID)
	if err != nil {
		return nil, errs.ToResolverErr(err)
	}

	if appVersionNumber == nil {
		return nil, nil
	}

	appVersion, err := t.deps.appService.FindAppVersionByAppIDAndNumber(ctx, t.teamAppInstallation.AppID, *appVersionNumber)
	if err != nil {
		return nil, errs.ToResolverErr(err)
	}

	gqlAppVersion := newAppVersion(t.deps, appVersion)
	return &gqlAppVersion, nil
}

func newTeamAppInstallation(
	deps *Dependencies,
	teamAppInstallation entity.TeamAppInstallation,
) TeamAppInstallation {
	return TeamAppInstallation{deps: deps, teamAppInstallation: teamAppInstallation}
}
