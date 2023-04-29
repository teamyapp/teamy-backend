//go:build wireinject

package dep

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/graph-gophers/graphql-go/trace/tracer"
	"github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/env"
	cloudGQL "github.com/teamyapp/cloud/libs/gql"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/api"
	"github.com/teamyapp/teamy-backend/core/api/gql"
	"github.com/teamyapp/teamy-backend/core/cache"
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
	wire.Bind(new(daov2.App), new(sqldbV2.App)),
	wire.Bind(new(daov2.AppVersion), new(sqldbV2.AppVersion)),
	wire.Bind(new(daov2.AppVersionVisibleTeam), new(sqldbV2.AppVersionVisibleTeam)),
	wire.Bind(new(daov2.AppTeamInstallation), new(sqldbV2.AppTeamInstallation)),
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
	sqldbV2.NewApp,
	sqldbV2.NewAppVersion,
	sqldbV2.NewAppVersionVisibleTeam,
	sqldbV2.NewAppTeamInstallation,
)

var serviceSet = wire.NewSet(
	service.NewThread,
	service.NewTask,
	service.NewTaskLink,
	service.NewInvitation,
	newTeamService,
	service.NewSprint,
	newUserService,
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
	cloudAPIClientRegistry *client.Registry,
	realTimeStateSyncer *realtime.StateSyncer,
	sqlDB *sql.DB,
) (cloudGQL.Service[gql.Resolver], error) {
	wire.Build(
		wire.Bind(new(tracer.Tracer), new(cloudGQL.PrometheusTracer)),

		newPrometheusTracer,
		daoSet,
		transaction.NewFactory,
		serviceSet,
		client.NewAuthorizer,
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
	cloudAPIClientRegistry *client.Registry,
	realTimeStateSyncer *realtime.StateSyncer,
	sqlDB *sql.DB,
) api.TaskRPC {
	wire.Build(
		daoSet,
		serviceSet,
		client.NewAuthorizer,
		feature.NewStaticToggles,
		cache.NewActivity,
		transaction.NewFactory,
		api.NewTaskRPC,
	)
	return api.TaskRPC{}
}

func InitSprintRPCAPI(
	logger telemetry.Logger,
	cloudAPIClientRegistry *client.Registry,
	realTimeStateSyncer *realtime.StateSyncer,
	sqlDB *sql.DB,
) api.SprintRPC {
	wire.Build(
		daoSet,
		serviceSet,
		client.NewAuthorizer,
		feature.NewStaticToggles,
		transaction.NewFactory,
		api.NewSprintRPC,
	)
	return api.SprintRPC{}
}

func InitTaskLinkRPCAPI(
	logger telemetry.Logger,
	cloudAPIClientRegistry *client.Registry,
	realTimeStateSyncer *realtime.StateSyncer,
	sqlDB *sql.DB,
) api.TaskLinkRPC {
	wire.Build(
		daoSet,
		serviceSet,
		client.NewAuthorizer,
		feature.NewStaticToggles,
		transaction.NewFactory,
		api.NewTaskLinkRPC,
	)
	return api.TaskLinkRPC{}
}

func newUserService(
	logger telemetry.Logger,
	cloudWebAPIExternalBaseURL CloudWebAPIExternalBaseURL,
	cloudClientRegistry *client.Registry,
	stateSyncer *realtime.StateSyncer,
	transactionFactory transaction.Factory,
	toggles feature.Toggles,
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
		userDaoV2,
		teamMemberV2,
		userFileUploadSessionDaoV2,
	)
}

func newTeamService(
	logger telemetry.Logger,
	cloudWebAPIExternalBaseURL CloudWebAPIExternalBaseURL,
	cloudClientRegistry *client.Registry,
	authorizer client.Authorizer,
	toggles feature.Toggles,
	stateSyncer *realtime.StateSyncer,
	transactionFactory transaction.Factory,
	taskDaoV2 daov2.Task,
	sprintDaoV2 daov2.Sprint,
	sprintParticipantDaoV2 daov2.SprintParticipant,
	teamDaoV2 daov2.Team,
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
		sprintDaoV2,
		sprintParticipantDaoV2,
		teamDaoV2,
		teamMemberDaoV2,
		teamFileUploadSessionDaoV2,
		teamGroupDaoV2)
}

func newPrometheusTracer(appMame AppMame, serviceName ServiceName, environment env.Environment) cloudGQL.PrometheusTracer {
	return cloudGQL.NewPrometheusTracer(string(appMame), string(serviceName), environment)
}
