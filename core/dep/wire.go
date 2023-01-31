//go:build wireinject

package dep

import (
	"database/sql"

	"github.com/google/wire"
	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/api"
	"github.com/teamyapp/teamy-backend/core/api/gql"
	"github.com/teamyapp/teamy-backend/core/cache"
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
	wire.Bind(new(dao.TaskLink), new(sqldb.TaskLink)),
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
	sqldb.NewTaskLink,
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

var serviceSet = wire.NewSet(
	service.NewThread,
	service.NewTask,
	service.NewTaskLink,
	service.NewInvitation,
	newTeamService,
	service.NewSprint,
	newUserService,
	service.NewAuthorizer,
)

func InitRealTimeStateSyncer(dataCollector telemetry.DataCollector, qlDB *sql.DB) *realtime.StateSyncer {
	wire.Build(
		daoSet,
		realtime.NewStateSyncer,
	)
	return nil
}

func InitGraphQLAPI(
	dataCollector telemetry.DataCollector,
	cloudWebAPIExternalBaseURL CloudWebAPIExternalBaseURL,
	cloudAPIClientRegistry *cloudAPI.ClientRegistry,
	realTimeStateSyncer *realtime.StateSyncer,
	sqlDB *sql.DB,
) (api.GraphQL, error) {
	wire.Build(
		daoSet,
		serviceSet,
		cache.NewActivity,
		gql.NewDependencies,
		gql.NewResolver,
		api.NewGraphQL,
	)
	return api.GraphQL{}, nil
}

func InitRealTimeStateSyncAPI(
	dataCollector telemetry.DataCollector,
	realTimeStateSyncer *realtime.StateSyncer,
) api.RealTimeStateSync {
	wire.Build(
		api.NewRealTimeStateSync,
	)
	return api.RealTimeStateSync{}
}

func InitTaskRPCAPI(
	dataCollector telemetry.DataCollector,
	cloudAPIClientRegistry *cloudAPI.ClientRegistry,
	realTimeStateSyncer *realtime.StateSyncer,
	sqlDB *sql.DB,
) api.TaskRPC {
	wire.Build(
		daoSet,
		cache.NewActivity,
		serviceSet,
		api.NewTaskRPC,
	)
	return api.TaskRPC{}
}

func InitSprintRPCAPI(
	dataCollector telemetry.DataCollector,
	cloudAPIClientRegistry *cloudAPI.ClientRegistry,
	realTimeStateSyncer *realtime.StateSyncer,
	sqlDB *sql.DB,
) api.SprintRPC {
	wire.Build(
		daoSet,
		serviceSet,
		cache.NewActivity,
		api.NewSprintRPC,
	)
	return api.SprintRPC{}
}

func InitTaskLinkRPCAPI(
	dataCollector telemetry.DataCollector,
	cloudAPIClientRegistry *cloudAPI.ClientRegistry,
	realTimeStateSyncer *realtime.StateSyncer,
	sqlDB *sql.DB,
) api.TaskLinkRPC {
	wire.Build(
		daoSet,
		serviceSet,
		api.NewTaskLinkRPC,
	)
	return api.TaskLinkRPC{}
}

func newUserService(
	dataCollector telemetry.DataCollector,
	cloudWebAPIExternalBaseURL CloudWebAPIExternalBaseURL,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	stateSyncer *realtime.StateSyncer,
	userDao dao.User,
	userFileUploadSessionDao dao.UserFileUploadSession,
	teamMemberDao dao.TeamMember,
) service.User {
	return service.NewUser(
		dataCollector,
		string(cloudWebAPIExternalBaseURL),
		cloudClientRegistry,
		stateSyncer,
		userDao,
		userFileUploadSessionDao,
		teamMemberDao,
	)
}

func newTeamService(
	dataCollector telemetry.DataCollector,
	cloudWebAPIExternalBaseURL CloudWebAPIExternalBaseURL,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	authorizer service.Authorizer,
	stateSyncer *realtime.StateSyncer,
	taskDao dao.Task,
	sprintDao dao.Sprint,
	teamDao dao.Team,
	teamMemberDao dao.TeamMember,
	teamFileUploadSessionDao dao.TeamFileUploadSession,
	sprintService service.Sprint,
) service.Team {
	return service.NewTeam(
		dataCollector,
		string(cloudWebAPIExternalBaseURL),
		cloudClientRegistry,
		authorizer,
		stateSyncer,
		taskDao,
		sprintDao,
		teamDao,
		teamMemberDao,
		teamFileUploadSessionDao,
		sprintService)
}
