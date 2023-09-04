package gql

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamAppInstallation struct {
	deps                *Dependencies
	teamAppInstallation entity.TeamAppInstallation
}

func (t TeamAppInstallation) InstalledTeam(ctx context.Context) (Team, error) {
	teamEntity, err := t.deps.teamService.FindTeamByID(ctx, t.teamAppInstallation.InstalledTeamID)
	if err != nil {
		t.deps.logger.ErrorWithContext(ctx, err)
		return Team{}, errs.ToResolverErr(err)
	}

	return newTeam(t.deps, teamEntity), nil
}

func (t TeamAppInstallation) ActiveAppVersion(ctx context.Context) (*AppVersion, *errs.Error) {
	appVersionNumber, err := t.deps.rolloutService.GetActiveAppVersionNumberForTeam(ctx, t.teamAppInstallation)
	if err != nil {
		return nil, err
	}

	appVersionEntity, err := t.deps.appService.FindAppVersionByAppIDAndNumber(ctx, t.teamAppInstallation.AppID, appVersionNumber)
	if err != nil {
		return nil, err
	}

	appVersion := newAppVersion(t.deps, appVersionEntity)
	return &appVersion, nil
}

func newTeamAppInstallation(deps *Dependencies, teamAppInstallation entity.TeamAppInstallation) TeamAppInstallation {
	return TeamAppInstallation{deps: deps, teamAppInstallation: teamAppInstallation}
}
