//go:build wireinject

package dep

import (
	"database/sql"

	"github.com/google/wire"
	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/api"
	"github.com/teamyapp/teamy-backend/core/api/gql"
	"github.com/teamyapp/teamy-backend/core/cache"
	"github.com/teamyapp/teamy-backend/core/collection"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/dao/sqldb"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/service"
)

type CloudWebAPIExternalBaseURL string

var daoSet = wire.NewSet(
	wire.Bind(new(dao.Invitation), new(sqldb.Invitation)),
	wire.Bind(new(dao.Message), new(sqldb.Message)),
	wire.Bind(new(dao.Task), new(sqldb.Task)),
	wire.Bind(new(dao.Team), new(sqldb.Team)),
	wire.Bind(new(dao.TeamMember), new(sqldb.TeamMember)),
	wire.Bind(new(dao.User), new(sqldb.User)),
	wire.Bind(new(dao.Thread), new(sqldb.Thread)),
	wire.Bind(new(dao.Sprint), new(sqldb.Sprint)),
	wire.Bind(new(dao.TaskAwaitForRelation), new(sqldb.TaskAwaitForRelation)),
	wire.Bind(new(dao.SprintTaskRelation), new(sqldb.SprintTaskRelation)),
	wire.Bind(new(dao.UserFileUploadSession), new(sqldb.UserFileUploadSession)),
	wire.Bind(new(dao.TeamFileUploadSession), new(sqldb.TeamFileUploadSession)),
	wire.Bind(new(dao.SprintParticipant), new(sqldb.SprintParticipant)),
	sqldb.NewInvitation,
	sqldb.NewMessage,
	sqldb.NewTask,
	sqldb.NewTeam,
	sqldb.NewTeamMember,
	sqldb.NewUser,
	sqldb.NewThread,
	sqldb.NewSprint,
	sqldb.NewTaskAwaitForRelation,
	sqldb.NewSprintTaskRelation,
	sqldb.NewUserFileUploadSession,
	sqldb.NewTeamFileUploadSession,
	sqldb.NewSprintParticipant,
)

var collectionSyncerSet = wire.NewSet(
	collection.NewInvitationSyncer,
	collection.NewMessageSyncer,
	collection.NewTaskSyncer,
	collection.NewTaskAwaitForRelationSyncer,
	collection.NewSprintTaskRelationSyncer,
	collection.NewTeamSyncer,
	collection.NewTeamMemberSyncer,
	collection.NewUserSyncer,
)

var serviceSet = wire.NewSet(
	service.NewThread,
	service.NewTask,
	newTeamService,
	service.NewSprint,
	newUserService,
)

func InitDataCollector(severity obs.Severity) obs.DataCollector {
	wire.Build(
		wire.Bind(new(obs.Logger), new(obs.RawLogger)),
		obs.NewRawLogger,
		obs.NewDataCollector,
	)
	return obs.DataCollector{}
}

func InitRealTimeStateSyncer(sdataCollector obs.DataCollector, qlDB *sql.DB) *realtime.StateSyncer {
	wire.Build(
		daoSet,
		realtime.NewStateSyncer,
	)
	return nil
}

func InitGraphQLAPI(
	dataCollector obs.DataCollector,
	cloudWebAPIExternalBaseURL CloudWebAPIExternalBaseURL,
	cloudAPIClientRegistry *cloudAPI.ClientRegistry,
	realTimeStateSyncer *realtime.StateSyncer,
	sqlDB *sql.DB,
) (api.GraphQL, error) {
	wire.Build(
		daoSet,
		collectionSyncerSet,
		serviceSet,
		cache.NewActivity,
		gql.NewDependencies,
		gql.NewResolver,
		api.NewGraphQL,
	)
	return api.GraphQL{}, nil
}

func InitRealTimeStateSyncAPI(
	dataCollector obs.DataCollector,
	realTimeStateSyncer *realtime.StateSyncer,
) api.RealTimeStateSync {
	wire.Build(
		api.NewRealTimeStateSync,
	)
	return api.RealTimeStateSync{}
}

func InitTaskRPCAPI(
	dataCollector obs.DataCollector,
	cloudAPIClientRegistry *cloudAPI.ClientRegistry,
	realTimeStateSyncer *realtime.StateSyncer,
	sqlDB *sql.DB,
) api.TaskRPC {
	wire.Build(
		daoSet,
		collectionSyncerSet,
		cache.NewActivity,
		serviceSet,
		api.NewTaskRPC,
	)
	return api.TaskRPC{}
}

func InitSprintRPCAPI(
	dataCollector obs.DataCollector,
	cloudAPIClientRegistry *cloudAPI.ClientRegistry,
	realTimeStateSyncer *realtime.StateSyncer,
	sqlDB *sql.DB,
) api.SprintRPC {
	wire.Build(
		daoSet,
		collectionSyncerSet,
		serviceSet,
		api.NewSprintRPC,
	)
	return api.SprintRPC{}
}

func newUserService(
	dataCollector obs.DataCollector,
	cloudWebAPIExternalBaseURL CloudWebAPIExternalBaseURL,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	userDao dao.User,
	userFileUploadSessionDao dao.UserFileUploadSession,
) service.User {
	return service.NewUser(
		dataCollector,
		string(cloudWebAPIExternalBaseURL),
		cloudClientRegistry,
		userDao,
		userFileUploadSessionDao)
}

func newTeamService(
	dataCollector obs.DataCollector,
	cloudWebAPIExternalBaseURL CloudWebAPIExternalBaseURL,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	taskDao dao.Task,
	sprintDao dao.Sprint,
	teamDao dao.Team,
	teamFileUploadSessionDao dao.TeamFileUploadSession,
) service.Team {
	return service.NewTeam(
		dataCollector,
		string(cloudWebAPIExternalBaseURL),
		cloudClientRegistry,
		taskDao,
		sprintDao,
		teamDao,
		teamFileUploadSessionDao)
}
