//go:build wireinject

package dep

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/teamyapp/cloud/app/api/rpc"
	"github.com/teamyapp/teamy-backend/apps"
	appsConfig "github.com/teamyapp/teamy-backend/apps/config"
	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/dao/sqldb"
	"github.com/teamyapp/teamy-backend/config"
)

func InitGithubApp(
	cloudAPICfg config.CloudAPIConfig,
	config appsConfig.GithubAppConfig,
	sqlDB *sql.DB,
) (apps.GithubApp, error) {
	wire.Build(
		wire.Bind(new(dao.GithubAppInstallState), new(sqldb.GithubAppInstallState)),
		wire.Bind(new(dao.GithubAppInstallation), new(sqldb.GithubAppInstallation)),

		sqldb.NewGithubAppInstallState,
		sqldb.NewGithubAppInstallation,
		newClientAPIClient,
		apps.NewGithubApp,
	)
	return apps.GithubApp{}, nil
}

func newClientAPIClient(
	cfg config.CloudAPIConfig,
) (*rpc.CloudAPIClient, error) {
	return rpc.NewCloudAPIClient(cfg.Host, cfg.Port, cfg.ShouldEncrypt)
}
