//go:build wireinject

package dep

import (
	"database/sql"

	"github.com/google/wire"
	cloudClient "github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/gql"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/dao/sqldb"
	"github.com/teamyapp/teamy-backend/apps/github"
	"github.com/teamyapp/teamy-backend/apps/github/client"
	teamyClient "github.com/teamyapp/teamy-backend/core/client"
)

type GithubAppPrivateKeyPEM []byte
type TeamyWebUIBaseURL string

func InitGithubAppAPI(
	logger telemetry.Logger,
	cloudClientRegistry *cloudClient.Registry,
	teamyClientRegistry *teamyClient.Registry,
	httpClient web.HTTPClient,
	config github.AppConfig,
	githubAppPrivateKeyPEM GithubAppPrivateKeyPEM,
	sqlDB *sql.DB,
	teamyWebUIBaseURL TeamyWebUIBaseURL,
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
		newAppAPI,
	)
	return github.AppAPI{}, nil
}

func newAppAPI(
	cfg github.AppConfig,
	logger telemetry.Logger,
	cloudClientRegistry *cloudClient.Registry,
	teamyClientRegistry *teamyClient.Registry,
	githubAppInstallStateDao dao.GithubAppInstallState,
	githubAppInstallationDao dao.GithubAppInstallation,
	githubPullRequestDao dao.GithubPullRequest,
	githubCodeReviewDao dao.GithubCodeReview,
	githubRequiredUserActionDao dao.GithubRequiredUserAction,
	githubPullRequestInternalTaskRelationDao dao.GithubPullRequestInternalTaskRelation,
	githubGraphQLAPI client.GraphQLAPI,
	githubRESTAPI client.RESTAPI,
	githubApp *client.GithubApp,
	teamyWebUIBaseURL TeamyWebUIBaseURL,
) github.AppAPI {
	return github.NewAppAPI(
		cfg,
		logger,
		cloudClientRegistry,
		teamyClientRegistry,
		githubAppInstallStateDao,
		githubAppInstallationDao,
		githubPullRequestDao,
		githubCodeReviewDao,
		githubRequiredUserActionDao,
		githubPullRequestInternalTaskRelationDao,
		githubGraphQLAPI,
		githubRESTAPI,
		githubApp,
		string(teamyWebUIBaseURL),
	)
}

func newGithubApp(logger telemetry.Logger, config github.AppConfig, privateKeyPEM GithubAppPrivateKeyPEM) (*client.GithubApp, error) {
	return client.NewGithubApp(logger, config.AppID, []byte(privateKeyPEM))
}
