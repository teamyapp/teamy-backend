//go:build wireinject

package dep

import (
	"database/sql"

	"github.com/google/wire"
	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/teamy-backend/core/api"
	"github.com/teamyapp/teamy-backend/core/api/gql"
	"github.com/teamyapp/teamy-backend/core/collection"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/dao/sqldb"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/service"
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
)

var serviceSet = wire.NewSet(
	service.NewThread,
	service.NewTask,
)

func InitRealTimeStateSyncer(sqlDB *sql.DB) *realtime.StateSyncer {
	wire.Build(
		daoSet,
		realtime.NewStateSyncer,
	)
	return nil
}

func InitGraphQLAPI(
	cloudAPIClientRegistry *cloudAPI.ClientRegistry,
	realTimeStateSyncer *realtime.StateSyncer,
	sqlDB *sql.DB,
) (api.GraphQL, error) {
	wire.Build(
		daoSet,
		collectionSyncerSet,
		serviceSet,
		gql.NewDependencies,
		gql.NewResolver,
		api.NewGraphQL,
	)
	return api.GraphQL{}, nil
}

func InitRealTimeStateSyncAPI(
	realTimeStateSyncer *realtime.StateSyncer,
) api.RealTimeStateSync {
	wire.Build(
		api.NewRealTimeStateSync,
	)
	return api.RealTimeStateSync{}
}

func InitTaskRPCAPI(
	cloudAPIClientRegistry *cloudAPI.ClientRegistry,
	realTimeStateSyncer *realtime.StateSyncer,
	sqlDB *sql.DB,
) api.TaskRPC {
	wire.Build(
		daoSet,
		collectionSyncerSet,
		serviceSet,
		api.NewTaskRPC,
	)
	return api.TaskRPC{}
}
