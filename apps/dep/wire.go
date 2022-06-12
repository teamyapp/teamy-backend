//go:build wireinject

package dep

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/teamyapp/cloud/app/api/rpc"
	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/teamy-backend/apps"
	appsConfig "github.com/teamyapp/teamy-backend/apps/config"
	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/dao/sqldb"
)

func InitGithubApp(
	cloudAPICfg config.CloudAPIClient,
	config appsConfig.GithubAppConfig,
	sqlDB *sql.DB,
) (apps.GithubApp, error) {
	wire.Build(
		wire.Bind(new(dao.GithubAppInstallState), new(sqldb.GithubAppInstallState)),
		wire.Bind(new(dao.GithubAppInstallation), new(sqldb.GithubAppInstallation)),

		sqldb.NewGithubAppInstallState,
		sqldb.NewGithubAppInstallation,
		rpc.NewCloudAPIClient,
		apps.NewGithubApp,
	)
	return apps.GithubApp{}, nil
}
