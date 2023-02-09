//go:build wireinject

package dep

import (
	"database/sql"

	"github.com/google/wire"
	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/dao/sqldb"
	"github.com/teamyapp/teamy-backend/apps/github"
	"github.com/teamyapp/teamy-backend/core/api"
)

func InitGithubAppAPI(
	dataCollector telemetry.DataCollector,
	cloudAPIClientRegistry *cloudAPI.ClientRegistry,
	teamyAPIClientRegistry *api.ClientRegistry,
	config github.AppConfig,
	sqlDB *sql.DB,
) (github.AppAPI, error) {
	wire.Build(
		wire.Bind(new(dao.GithubAppInstallState), new(sqldb.GithubAppInstallState)),
		wire.Bind(new(dao.GithubAppInstallation), new(sqldb.GithubAppInstallation)),
		wire.Bind(new(dao.GithubPullRequest), new(sqldb.GithubPullRequest)),
		wire.Bind(new(dao.GithubCodeReview), new(sqldb.GithubCodeReview)),
		wire.Bind(new(dao.GithubRequiredUserAction), new(sqldb.GithubRequiredUserAction)),

		sqldb.NewGithubAppInstallState,
		sqldb.NewGithubAppInstallation,
		sqldb.NewGithubPullRequest,
		sqldb.NewGithubCodeReview,
		sqldb.NewGithubRequiredUserAction,
		github.NewAppAPI,
	)
	return github.AppAPI{}, nil
}
