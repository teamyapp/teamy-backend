package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppVersionChange struct {
	appVersionChange entity.AppVersionChange
	deps             *Dependencies
}

func (a AppVersionChange) ID() graphql.ID {
	return toGraphQLID(a.appVersionChange.ID)
}

func (a AppVersionChange) AppVersion(ctx context.Context) (AppVersion, error) {
	appVersion, err := a.deps.appService.FindAppVersionByAppIDAndNumber(ctx, a.appVersionChange.AppID, a.appVersionChange.VersionNumber)
	if err != nil {
		a.deps.logger.ErrorWithContext(ctx, err)
		return AppVersion{}, errs.ToResolverErr(err)
	}

	return newAppVersion(a.deps, appVersion), nil
}

func (a AppVersionChange) Change() string {
	return a.appVersionChange.Change
}

func newAppVersionChange(deps *Dependencies, appVersionChange entity.AppVersionChange) AppVersionChange {
	return AppVersionChange{
		appVersionChange: appVersionChange,
		deps:             deps,
	}
}
