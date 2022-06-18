//go:build wireinject

package dep

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/teamyapp/cloud/app/api/rpc"
	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/teamy-backend/core/api/gql/resolver"
	"github.com/teamyapp/teamy-backend/core/collection"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/dao/sqldb"
	"github.com/teamyapp/teamy-backend/infras/storage"
)

var daoSet = wire.NewSet(
	wire.Bind(new(dao.Invitation), new(sqldb.Invitation)),
	wire.Bind(new(dao.Message), new(sqldb.Message)),
	wire.Bind(new(dao.Task), new(sqldb.Task)),
	wire.Bind(new(dao.TaskAwaitForRelation), new(sqldb.TaskAwaitForRelation)),
	wire.Bind(new(dao.Team), new(sqldb.Team)),
	wire.Bind(new(dao.TeamMember), new(sqldb.TeamMember)),
	wire.Bind(new(dao.User), new(sqldb.User)),
	wire.Bind(new(dao.Thread), new(sqldb.Thread)),
	sqldb.NewInvitation,
	sqldb.NewMessage,
	sqldb.NewTask,
	sqldb.NewTaskAwaitForRelation,
	sqldb.NewTeam,
	sqldb.NewTeamMember,
	sqldb.NewUser,
	sqldb.NewThread,
)

var collectionSyncerSet = wire.NewSet(
	collection.NewInvitationSyncer,
	collection.NewMessageSyncer,
	collection.NewTaskSyncer,
	collection.NewTaskAwaitForRelationSyncer,
	collection.NewTeamSyncer,
	collection.NewTeamMemberSyncer,
	collection.NewUserSyncer,
	collection.NewThreadSyncer,
)

func InitGraphQLResolver(
	sqlDB *sql.DB,
	cloudAPIClientCfg config.CloudAPIClient,
	rtCollections *storage.RealTimeCollections,
) (resolver.Resolver, error) {
	wire.Build(
		daoSet,
		collectionSyncerSet,
		rpc.NewCloudAPIClient,
		resolver.NewDependencies,
		resolver.NewResolver,
	)
	return resolver.Resolver{}, nil
}
