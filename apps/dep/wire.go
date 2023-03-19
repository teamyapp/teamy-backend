//go:build wireinject

package dep

import (
	"database/sql"

	"github.com/google/wire"
	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/libs/gql"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/dao/sqldb"
	"github.com/teamyapp/teamy-backend/apps/github"
	"github.com/teamyapp/teamy-backend/apps/github/client"
	"github.com/teamyapp/teamy-backend/core/api"
)

type GithubAppPrivateKeyPEM []byte

func InitGithubAppAPI(
	dataCollector telemetry.DataCollector,
	cloudAPIClientRegistry *cloudAPI.ClientRegistry,
	teamyAPIClientRegistry *api.ClientRegistry,
	httpClient web.HTTPClient,
	config github.AppConfig,
	githubAppPrivateKeyPEM GithubAppPrivateKeyPEM,
	sqlDB *sql.DB,
) (github.AppAPI, error) {
	wire.Build(
		wire.Bind(new(dao.GithubAppInstallState), new(sqldb.GithubAppInstallState)),
		wire.Bind(new(dao.GithubAppInstallation), new(sqldb.GithubAppInstallation)),
		wire.Bind(new(dao.GithubPullRequest), new(sqldb.GithubPullRequest)),
		wire.Bind(new(dao.GithubCodeReview), new(sqldb.GithubCodeReview)),
		wire.Bind(new(dao.GithubRequiredUserAction), new(sqldb.GithubRequiredUserAction)),
		wire.Bind(new(dao.GithubPullRequestInternalTaskRelation), new(sqldb.GithubPullRequestInternalTaskRelation)),
		sqldb.NewGithubAppInstallState,
		sqldb.NewGithubAppInstallation,
		sqldb.NewGithubPullRequest,
		sqldb.NewGithubCodeReview,
		sqldb.NewGithubRequiredUserAction,
		sqldb.NewGithubPullRequestInternalTaskRelation,
		newGithubApp,
		gql.NewClient,
		client.NewGraphQLAPI,
		client.NewRESTAPI,
		github.NewAppAPI,
	)
	return github.AppAPI{}, nil
}

func newGithubApp(dataCollector telemetry.DataCollector, config github.AppConfig, privateKeyPEM GithubAppPrivateKeyPEM) (*client.GithubApp, error) {
	return client.NewGithubApp(dataCollector, config.AppID, []byte(privateKeyPEM))
}
