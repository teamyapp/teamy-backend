//go:build wireinject

package dep

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/config"
	appsConfig "github.com/teamyapp/teamy-backend/apps/config"
	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/dao/sqldb"
	"github.com/teamyapp/teamy-backend/apps/github"
)

func InitGithubApp(
	cloudAPICfg config.CloudAPIClient,
	config appsConfig.GithubAppConfig,
	sqlDB *sql.DB,
) (github.App, error) {
	wire.Build(
		wire.Bind(new(dao.GithubAppInstallState), new(sqldb.GithubAppInstallState)),
		wire.Bind(new(dao.GithubAppInstallation), new(sqldb.GithubAppInstallation)),

		sqldb.NewGithubAppInstallState,
		sqldb.NewGithubAppInstallation,
		api.NewCloudAPIClient,
		github.NewApp,
	)
	return github.App{}, nil
}
