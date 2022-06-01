//go:build wireinject

package dep

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/teamyapp/cloud/app/api/rpc"
	"github.com/teamyapp/teamy-backend/app/api/gql/resolver"
	"github.com/teamyapp/teamy-backend/app/dao"
	"github.com/teamyapp/teamy-backend/app/dao/sqldb"
)

var daoSet = wire.NewSet(
	wire.Bind(new(dao.Invitation), new(sqldb.Invitation)),
	wire.Bind(new(dao.Message), new(sqldb.Message)),
	wire.Bind(new(dao.Task), new(sqldb.Task)),
	wire.Bind(new(dao.Team), new(sqldb.Team)),
	wire.Bind(new(dao.TeamMember), new(sqldb.TeamMember)),
	wire.Bind(new(dao.User), new(sqldb.User)),
	wire.Bind(new(dao.Thread), new(sqldb.Thread)),
	sqldb.NewInvitation,
	sqldb.NewMessage,
	sqldb.NewTask,
	sqldb.NewTeam,
	sqldb.NewTeamMember,
	sqldb.NewUser,
	sqldb.NewThread,
)

func InitGraphQLResolver(
	sqlDB *sql.DB,
	cloudAPIHost CloudAPIHost,
	cloudAPIPort CloudAPIPort,
	cloudAPIShouldEncrypt CloudAPIShouldEncrypt,
) (resolver.Resolver, error) {
	wire.Build(
		daoSet,
		newClientAPIClient,
		resolver.NewDependencies,
		resolver.NewResolver,
	)
	return resolver.Resolver{}, nil
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
