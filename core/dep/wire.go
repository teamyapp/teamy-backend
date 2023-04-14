//go:build wireinject

package dep

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/graph-gophers/graphql-go/trace/tracer"
	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/libs/env"
	cloudGQL "github.com/teamyapp/cloud/libs/gql"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/api"
	"github.com/teamyapp/teamy-backend/core/api/gql"
	"github.com/teamyapp/teamy-backend/core/cache"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/dao/sqldb"
	"github.com/teamyapp/teamy-backend/core/daov2"
	sqldbV2 "github.com/teamyapp/teamy-backend/core/daov2/sqldb"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/service"
)

type AppMame string
type ServiceName string
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
	wire.Bind(new(dao.AppTeamInstallation), new(sqldb.AppTeamInstallation)),
	wire.Bind(new(dao.AppVersion), new(sqldb.AppVersion)),
	wire.Bind(new(dao.AppVersionVisibleTeam), new(sqldb.AppVersionVisibleTeam)),
	wire.Bind(new(dao.App), new(sqldb.App)),
	wire.Bind(new(daov2.Task), new(sqldbV2.Task)),
	wire.Bind(new(daov2.TaskLink), new(sqldbV2.TaskLink)),
	wire.Bind(new(daov2.TaskAwaitForRelation), new(sqldbV2.TaskAwaitForRelation)),
	wire.Bind(new(daov2.SprintParticipant), new(sqldbV2.SprintParticipant)),
	wire.Bind(new(daov2.Sprint), new(sqldbV2.Sprint)),
	wire.Bind(new(daov2.SprintTaskRelation), new(sqldbV2.SprintTaskRelation)),
	wire.Bind(new(daov2.Thread), new(sqldbV2.Thread)),
	wire.Bind(new(daov2.TeamMember), new(sqldbV2.TeamMember)),
	wire.Bind(new(daov2.TeamGroup), new(sqldbV2.TeamGroup)),
	wire.Bind(new(daov2.User), new(sqldbV2.User)),
	wire.Bind(new(daov2.UserFileUploadSession), new(sqldbV2.UserFileUploadSession)),
	wire.Bind(new(daov2.Team), new(sqldbV2.Team)),
	wire.Bind(new(daov2.TeamFileUploadSession), new(sqldbV2.TeamFileUploadSession)),
	wire.Bind(new(daov2.Invitation), new(sqldbV2.Invitation)),
	wire.Bind(new(daov2.Message), new(sqldbV2.Message)),
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
	sqldb.NewAppTeamInstallation,
	sqldb.NewAppVersion,
	sqldb.NewAppVersionVisibleTeam,
	sqldb.NewApp,
	sqldbV2.NewTask,
	sqldbV2.NewTaskLink,
	sqldbV2.NewTaskAwaitForRelation,
	sqldbV2.NewSprintParticipant,
	sqldbV2.NewSprint,
	sqldbV2.NewSprintTaskRelation,
	sqldbV2.NewThread,
	sqldbV2.NewTeamMember,
	sqldbV2.NewTeamGroup,
	sqldbV2.NewUser,
	sqldbV2.NewUserFileUploadSession,
	sqldbV2.NewTeam,
	sqldbV2.NewTeamFileUploadSession,
	sqldbV2.NewInvitation,
	sqldbV2.NewMessage,
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
	service.NewApp,
)

func InitRealTimeStateSyncer(logger telemetry.Logger, sqlDB *sql.DB) *realtime.StateSyncer {
	wire.Build(
		daoSet,
		realtime.NewStateSyncer,
		transaction.NewFactory,
	)
	return nil
}

func InitGraphQLAPI(
	appName AppMame,
	serviceName ServiceName,
	environment env.Environment,
	logger telemetry.Logger,
	cloudWebAPIExternalBaseURL CloudWebAPIExternalBaseURL,
	cloudAPIClientRegistry *cloudAPI.ClientRegistry,
	realTimeStateSyncer *realtime.StateSyncer,
	sqlDB *sql.DB,
) (cloudGQL.Service[gql.Resolver], error) {
	wire.Build(
		wire.Bind(new(tracer.Tracer), new(cloudGQL.PrometheusTracer)),

		newPrometheusTracer,
		daoSet,
		transaction.NewFactory,
		serviceSet,
		feature.NewStaticToggles,
		cache.NewActivity,
		gql.NewDependencies,
		gql.NewResolver,
		api.NewGraphQL,
	)
	return cloudGQL.Service[gql.Resolver]{}, nil
}

func InitRealTimeStateSyncAPI(
	logger telemetry.Logger,
	realTimeStateSyncer *realtime.StateSyncer,
) api.RealTimeStateSync {
	wire.Build(
		api.NewRealTimeStateSync,
	)
	return api.RealTimeStateSync{}
}

func InitTaskRPCAPI(
	logger telemetry.Logger,
	cloudAPIClientRegistry *cloudAPI.ClientRegistry,
	realTimeStateSyncer *realtime.StateSyncer,
	sqlDB *sql.DB,
) api.TaskRPC {
	wire.Build(
		daoSet,
		serviceSet,
		feature.NewStaticToggles,
		cache.NewActivity,
		transaction.NewFactory,
		api.NewTaskRPC,
	)
	return api.TaskRPC{}
}

func InitSprintRPCAPI(
	logger telemetry.Logger,
	cloudAPIClientRegistry *cloudAPI.ClientRegistry,
	realTimeStateSyncer *realtime.StateSyncer,
	sqlDB *sql.DB,
) api.SprintRPC {
	wire.Build(
		daoSet,
		serviceSet,
		feature.NewStaticToggles,
		transaction.NewFactory,
		api.NewSprintRPC,
	)
	return api.SprintRPC{}
}

func InitTaskLinkRPCAPI(
	logger telemetry.Logger,
	cloudAPIClientRegistry *cloudAPI.ClientRegistry,
	realTimeStateSyncer *realtime.StateSyncer,
	sqlDB *sql.DB,
) api.TaskLinkRPC {
	wire.Build(
		daoSet,
		serviceSet,
		feature.NewStaticToggles,
		transaction.NewFactory,
		api.NewTaskLinkRPC,
	)
	return api.TaskLinkRPC{}
}

func newUserService(
	logger telemetry.Logger,
	cloudWebAPIExternalBaseURL CloudWebAPIExternalBaseURL,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	stateSyncer *realtime.StateSyncer,
	transactionFactory transaction.Factory,
	toggles feature.Toggles,
	userDao dao.User,
	userDaoV2 daov2.User,
	teamMemberV2 daov2.TeamMember,
	userFileUploadSessionDaoV2 daov2.UserFileUploadSession,
) service.User {
	return service.NewUser(
		logger,
		string(cloudWebAPIExternalBaseURL),
		cloudClientRegistry,
		stateSyncer,
		transactionFactory,
		toggles,
		userDao,
		userDaoV2,
		teamMemberV2,
		userFileUploadSessionDaoV2,
	)
}

func newTeamService(
	logger telemetry.Logger,
	cloudWebAPIExternalBaseURL CloudWebAPIExternalBaseURL,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	authorizer service.Authorizer,
	toggles feature.Toggles,
	stateSyncer *realtime.StateSyncer,
	transactionFactory transaction.Factory,
	taskDaoV2 daov2.Task,
	sprintDao dao.Sprint,
	sprintDaoV2 daov2.Sprint,
	sprintParticipantDao dao.SprintParticipant,
	sprintParticipantDaoV2 daov2.SprintParticipant,
	teamDao dao.Team,
	teamDaoV2 daov2.Team,
	teamMemberDao dao.TeamMember,
	teamMemberDaoV2 daov2.TeamMember,
	teamFileUploadSessionDaoV2 daov2.TeamFileUploadSession,
	teamGroupDaoV2 daov2.TeamGroup,
) service.Team {
	return service.NewTeam(
		logger,
		string(cloudWebAPIExternalBaseURL),
		cloudClientRegistry,
		authorizer,
		toggles,
		stateSyncer,
		transactionFactory,
		taskDaoV2,
		sprintDao,
		sprintDaoV2,
		sprintParticipantDao,
		sprintParticipantDaoV2,
		teamDao,
		teamDaoV2,
		teamMemberDao,
		teamMemberDaoV2,
		teamFileUploadSessionDaoV2,
		teamGroupDaoV2)
}

func newPrometheusTracer(appMame AppMame, serviceName ServiceName, environment env.Environment) cloudGQL.PrometheusTracer {
	return cloudGQL.NewPrometheusTracer(string(appMame), string(serviceName), environment)
}
