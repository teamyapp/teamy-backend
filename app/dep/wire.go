//go:build wireinject

package dep

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/teamyapp/teamy-backend/app/api/gql/datastore"
	"github.com/teamyapp/teamy-backend/app/api/gql/resolver"
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
		wire.Bind(new(datastore.Persister), new(datastore.PostgresPersister)),
		service.NewPrioritization,
		resolver.NewResolver,
		resolver.NewDependencies,
		datastore.NewPostgresPersister,
		datastore.NewDataStore,
	)
	return resolver.Resolver{}
}
