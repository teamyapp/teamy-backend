//go:build wireinject

package dep

import (
	"database/sql"

	"github.com/google/wire"
	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/dao/sqldb"
	"github.com/teamyapp/teamy-backend/apps/github"
	"github.com/teamyapp/teamy-backend/core/api"
)

func InitGithubApp(
	cloudAPIClientRegistry *cloudAPI.ClientRegistry,
	teamyAPIClientRegistry *api.ClientRegistry,
	config github.AppConfig,
	sqlDB *sql.DB,
) (github.App, error) {
	wire.Build(
		wire.Bind(new(dao.GithubAppInstallState), new(sqldb.GithubAppInstallState)),
		wire.Bind(new(dao.GithubAppInstallation), new(sqldb.GithubAppInstallation)),

		sqldb.NewGithubAppInstallState,
		sqldb.NewGithubAppInstallation,
		github.NewApp,
	)
	return github.App{}, nil
}
