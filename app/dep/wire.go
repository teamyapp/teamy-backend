//go:build wireinject

package dep

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/teamyapp/cloud/app/api/rpc"
	"github.com/teamyapp/teamy-backend/app/api/gql/datastore"
	"github.com/teamyapp/teamy-backend/app/api/gql/resolver"
	resolver2 "github.com/teamyapp/teamy-backend/app/api/gqlv2/resolver"
	"github.com/teamyapp/teamy-backend/app/dao"
	"github.com/teamyapp/teamy-backend/app/dao/sqldb"
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

var daoSet = wire.NewSet(
	wire.Bind(new(dao.Invitation), new(sqldb.Invitation)),
	wire.Bind(new(dao.Message), new(sqldb.Message)),
	wire.Bind(new(dao.Task), new(sqldb.Task)),
	wire.Bind(new(dao.Team), new(sqldb.Team)),
	wire.Bind(new(dao.TeamMember), new(sqldb.TeamMember)),
	wire.Bind(new(dao.User), new(sqldb.User)),
	sqldb.NewInvitation,
	sqldb.NewMessage,
	sqldb.NewTask,
	sqldb.NewTeam,
	sqldb.NewTeamMember,
	sqldb.NewUser,
)

func InitGraphQLV2Resolver(
	sqlDB *sql.DB,
	cloudAPIHost CloudAPIHost,
	cloudAPIPort CloudAPIPort,
	cloudAPIShouldEncrypt CloudAPIShouldEncrypt,
) (resolver2.Resolver, error) {
	wire.Build(
		daoSet,
		newClientAPIClient,
		resolver2.NewDependencies,
		resolver2.NewResolver,
	)
	return resolver2.Resolver{}, nil
}

type CloudAPIHost string
type CloudAPIPort int
type CloudAPIShouldEncrypt bool

func newClientAPIClient(
	host CloudAPIHost,
	port CloudAPIPort,
	shouldEncrypt CloudAPIShouldEncrypt,
) (*rpc.CloudAPIClient, error) {
	return rpc.NewCloudAPIClient(string(host), int(port), bool(shouldEncrypt))
}
