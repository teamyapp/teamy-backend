//go:build wireinject

package dep

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/teamyapp/teamy-backend/app/api/gql/resolver"
	resolver2 "github.com/teamyapp/teamy-backend/app/api/gqlv2/resolver"
	"github.com/teamyapp/teamy-backend/app/repo"
	"github.com/teamyapp/teamy-backend/app/service"
)

var repoSet = wire.NewSet(
	wire.Bind(new(repo.Team), new(repo.SQLTeam)),
	wire.Bind(new(repo.Task), new(repo.SQLTask)),
	wire.Bind(new(repo.User), new(repo.SQLUser)),
	repo.NewSQLTeam,
	repo.NewSQLTask,
	repo.NewSQLUser,
)

func InitGraphQLResolver(sqlDB *sql.DB) resolver.Resolver {
	wire.Build(
		repoSet,
		service.NewPrioritization,
		service.NewTeam,
		service.NewTask,
		service.NewUser,
		service.NewExecution,
		resolver.NewResolver,
		resolver.NewDependencies,
		resolver2.NewDependencies,
	)
	return resolver.Resolver{}
}
