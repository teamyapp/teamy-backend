//go:build wireinject

package dep

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/graph-gophers/graphql-go/trace/tracer"
	"github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/env"
	cloudGQL "github.com/teamyapp/cloud/libs/gql"
	"github.com/teamyapp/cloud/libs/security"
	"github.com/teamyapp/cloud/libs/storage"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/api"
	"github.com/teamyapp/teamy-backend/core/api/gql"
	"github.com/teamyapp/teamy-backend/core/cache"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/dao/sqldb"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/repository"
	"github.com/teamyapp/teamy-backend/core/service"
)

type AppMame string
type JWTSigningKey string
type ServiceName string
type CloudWebAPIExternalBaseURL string
type MapServerURL string

var daoSet = wire.NewSet(
	wire.Bind(new(dao.Task), new(sqldb.Task)),
	wire.Bind(new(dao.TaskLink), new(sqldb.TaskLink)),
	wire.Bind(new(dao.TaskAwaitForRelation), new(sqldb.TaskAwaitForRelation)),
	wire.Bind(new(dao.SprintParticipant), new(sqldb.SprintParticipant)),
	wire.Bind(new(dao.Sprint), new(sqldb.Sprint)),
	wire.Bind(new(dao.SprintTaskRelation), new(sqldb.SprintTaskRelation)),
	wire.Bind(new(dao.Thread), new(sqldb.Thread)),
	wire.Bind(new(dao.TeamMember), new(sqldb.TeamMember)),
	wire.Bind(new(dao.TeamMemberGroup), new(sqldb.TeamMemberGroup)),
	wire.Bind(new(dao.TeamMemberGroupUserRelation), new(sqldb.TeamMemberGroupUserRelation)),
	wire.Bind(new(dao.User), new(sqldb.User)),
	wire.Bind(new(dao.UserFileUploadSession), new(sqldb.UserFileUploadSession)),
	wire.Bind(new(dao.Team), new(sqldb.Team)),
	wire.Bind(new(dao.TeamFileUploadSession), new(sqldb.TeamFileUploadSession)),
	wire.Bind(new(dao.Invitation), new(sqldb.Invitation)),
	wire.Bind(new(dao.Message), new(sqldb.Message)),
	wire.Bind(new(dao.AppPackageUploadSession), new(*sqldb.AppPackageUploadSession)),
	wire.Bind(new(dao.AppVersion), new(*sqldb.AppVersion)),
	wire.Bind(new(dao.App), new(*sqldb.App)),
	wire.Bind(new(dao.Tag), new(*sqldb.Tag)),
	wire.Bind(new(dao.AppTagRelation), new(*sqldb.AppTagRelation)),
	wire.Bind(new(dao.AppSecret), new(*sqldb.AppSecret)),
	wire.Bind(new(dao.TeamAppInstallation), new(*sqldb.TeamAppInstallation)),
	wire.Bind(new(dao.AppGroupRelation), new(*sqldb.AppGroupRelation)),
	wire.Bind(new(dao.AppRolloutRelation), new(*sqldb.AppRolloutRelation)),
	wire.Bind(new(dao.AppVersionChange), new(*sqldb.AppVersionChange)),
	wire.Bind(new(dao.AppVersionPrice), new(*sqldb.AppVersionPrice)),
	wire.Bind(new(dao.FilterGroup), new(*sqldb.FilterGroup)),
	wire.Bind(new(dao.GroupRolloutRelation), new(*sqldb.GroupRolloutRelation)),
	wire.Bind(new(dao.MaxViewersActivator), new(*sqldb.MaxViewersActivator)),
	wire.Bind(new(dao.PercentageActivator), new(*sqldb.PercentageActivator)),
	wire.Bind(new(dao.RolloutViewer), new(*sqldb.RolloutViewer)),
	wire.Bind(new(dao.Rollout), new(*sqldb.Rollout)),
	wire.Bind(new(dao.Group), new(*sqldb.Group)),
	wire.Bind(new(dao.TimeRangeActivator), new(*sqldb.TimeRangeActivator)),
	wire.Bind(new(dao.VersionSelectorVersionRelation), new(*sqldb.VersionSelectorVersionRelation)),
	wire.Bind(new(dao.VersionSelector), new(*sqldb.VersionSelector)),
	wire.Bind(new(dao.GroupMemberRelation), new(*sqldb.GroupMemberRelation)),
	wire.Bind(new(dao.Activator), new(*sqldb.Activator)),
	sqldb.NewTask,
	sqldb.NewTaskLink,
	sqldb.NewTaskAwaitForRelation,
	sqldb.NewSprintParticipant,
	sqldb.NewSprint,
	sqldb.NewSprintTaskRelation,
	sqldb.NewThread,
	sqldb.NewTeamMember,
	sqldb.NewTeamMemberGroup,
	sqldb.NewTeamMemberGroupUserRelation,
	sqldb.NewUser,
	sqldb.NewUserFileUploadSession,
	sqldb.NewTeam,
	sqldb.NewTeamFileUploadSession,
	sqldb.NewInvitation,
	sqldb.NewMessage,
	sqldb.NewAppPackageUploadSession,
	sqldb.NewAppVersion,
	sqldb.NewApp,
	sqldb.NewAppSecret,
	sqldb.NewTeamAppInstallation,
	sqldb.NewAppGroupRelation,
	sqldb.NewAppRolloutRelation,
	sqldb.NewAppVersionChange,
	sqldb.NewAppVersionPrice,
	sqldb.NewFilterGroup,
	sqldb.NewGroupRolloutRelation,
	sqldb.NewMaxViewersActivator,
	sqldb.NewPercentageActivator,
	sqldb.NewRolloutViewer,
	sqldb.NewRollout,
	sqldb.NewGroup,
	sqldb.NewTimeRangeActivator,
	sqldb.NewVersionSelectorVersionRelation,
	sqldb.NewVersionSelector,
	sqldb.NewTag,
	sqldb.NewAppTagRelation,
	sqldb.NewGroupMemberRelation,
	sqldb.NewActivator,
)

var repositorySet = wire.NewSet(
	repository.NewGroup,
	repository.NewActivator,
	repository.NewVersionSelector,
	repository.NewTeamMemberGroup,
)

var serviceSet = wire.NewSet(
	wire.Bind(new(storage.MapClient), new(*storage.HTTPClient)),
	newHTTPClient,
	service.NewThread,
	service.NewTask,
	service.NewTaskLink,
	service.NewInvitation,
	newTeamService,
	service.NewSprint,
	newUserService,
	service.NewApp,
	service.NewGroup,
	service.NewRollout,
)

func newJWTAuthority(logger telemetry.Logger, signingKey JWTSigningKey) security.JWTAuthority {
	return security.NewJWTAuthority(logger, string(signingKey))
}

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
	mapServerURL MapServerURL,
	cloudAPIClientRegistry *client.Registry,
	realTimeStateSyncer *realtime.StateSyncer,
	jwtSigningKey JWTSigningKey,
	sqlDB *sql.DB,
) (cloudGQL.Service[gql.Resolver], error) {
	wire.Build(
		wire.Bind(new(tracer.Tracer), new(cloudGQL.PrometheusTracer)),
		newPrometheusTracer,
		daoSet,
		repositorySet,
		transaction.NewFactory,
		serviceSet,
		client.NewAuthorizer,
		feature.NewStaticToggles,
		cache.NewActivity,
		gql.NewDependencies,
		gql.NewResolver,
		api.NewGraphQL,
		newJWTAuthority,
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

func InitTeamRPCAPI(
	logger telemetry.Logger,
	cloudAPIClientRegistry *client.Registry,
	realTimeStateSyncer *realtime.StateSyncer,
	sqlDB *sql.DB,
	cloudWebAPIExternalBaseURL CloudWebAPIExternalBaseURL,
) api.TeamRPC {
	wire.Build(
		daoSet,
		repositorySet,
		serviceSet,
		client.NewAuthorizer,
		feature.NewStaticToggles,
		transaction.NewFactory,
		api.NewTeamRPC,
	)
	return api.TeamRPC{}
}

func newHTTPClient(
	mapServerURL MapServerURL,
) *storage.HTTPClient {
	return storage.NewHTTPClient(string(mapServerURL))
}

func newUserService(
	logger telemetry.Logger,
	toggles feature.Toggles,
	cloudWebAPIExternalBaseURL CloudWebAPIExternalBaseURL,
	cloudClientRegistry *client.Registry,
	authorizer client.Authorizer,
	stateSyncer *realtime.StateSyncer,
	transactionFactory transaction.Factory,
	userDao dao.User,
	teamMember dao.TeamMember,
	userFileUploadSessionDao dao.UserFileUploadSession,
) service.User {
	return service.NewUser(
		logger,
		toggles,
		string(cloudWebAPIExternalBaseURL),
		cloudClientRegistry,
		authorizer,
		stateSyncer,
		transactionFactory,
		userDao,
		teamMember,
		userFileUploadSessionDao,
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
	taskDao dao.Task,
	sprintDao dao.Sprint,
	sprintParticipantDao dao.SprintParticipant,
	teamDao dao.Team,
	teamMemberDao dao.TeamMember,
	teamFileUploadSessionDao dao.TeamFileUploadSession,
	teamMemberGroupDao dao.TeamMemberGroup,
	teamMemberGroupUserRelationDao dao.TeamMemberGroupUserRelation,
	teamMemberGroupRepo repository.TeamMemberGroup,
) service.Team {
	return service.NewTeam(
		logger,
		string(cloudWebAPIExternalBaseURL),
		cloudClientRegistry,
		authorizer,
		toggles,
		stateSyncer,
		transactionFactory,
		taskDao,
		sprintDao,
		sprintParticipantDao,
		teamDao,
		teamMemberDao,
		teamFileUploadSessionDao,
		teamMemberGroupDao,
		teamMemberGroupUserRelationDao,
		teamMemberGroupRepo)
}

func newPrometheusTracer(appMame AppMame, serviceName ServiceName, environment env.Environment) cloudGQL.PrometheusTracer {
	return cloudGQL.NewPrometheusTracer(string(appMame), string(serviceName), environment)
}
