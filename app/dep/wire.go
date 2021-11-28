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
	repo.NewSQLTeam,
	repo.NewSQLTask,
)

func InitGraphQLResolver(sqlDB *sql.DB) resolver.Resolver {
	wire.Build(
		repoSet,
		service.NewPrioritization,
		service.NewTeam,
		service.NewTask,
		service.NewExecution,
		resolver.NewQuery,
		resolver.NewMutation,
		resolver.NewResolver,
		resolver2.NewDependencies,
	)
	return resolver.Resolver{}
}
