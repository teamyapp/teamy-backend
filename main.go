package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/cloud/libs/env"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/metrics"
	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/cloud/libs/network"
	"github.com/teamyapp/cloud/libs/retry"
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/rpc"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/runtime"
	"github.com/teamyapp/cloud/libs/telemetry"
	appsDep "github.com/teamyapp/teamy-backend/apps/dep"
	"github.com/teamyapp/teamy-backend/apps/github"
	appsDI "github.com/teamyapp/teamy-backend/apps/inject"
	"github.com/teamyapp/teamy-backend/config"
	"github.com/teamyapp/teamy-backend/core/api"
	"github.com/teamyapp/teamy-backend/core/dep"
	"github.com/teamyapp/teamy-backend/core/inject"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

const appName = "teamy"
const serviceName = "backend"

var serviceLabels = []string{appName, serviceName}
var fullServiceName = strings.Join(serviceLabels, "-")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	cfg, err := config.AppFromEnv()
	if err != nil {
		panic(err)
	}

	lineFormatter := newLineFormatter(cfg.Environment)
	logFileName := fmt.Sprintf("%v.log", fullServiceName)
	logFilePath := getEnv("LOG_OUTPUT_FILE", filepath.Join("..", "logs", logFileName))
	logOutput, err := telemetry.NewLogOutput(cfg.Environment, logFilePath)
	if err != nil {
		panic(err)
	}

	defer logOutput.Close()

	serviceLabelsWithEnv := append([]string{}, serviceLabels...)
	serviceLabelsWithEnv = append(serviceLabelsWithEnv, strings.ToLower(string(cfg.Environment)))
	logger := telemetry.NewLogger(
		lineFormatter,
		logOutput,
		cfg.LogVisibleLevel,
		[]telemetry.LogInterceptor{
			telemetry.NewCommitLogInterceptor(cfg.GitLongCommitHash),
			telemetry.NewServiceLogInterceptor(strings.Join(serviceLabelsWithEnv, "/")),
			telemetry.TraceLogInterceptor,
			telemetry.RequestLogInterceptor,
			realtime.MutationLogInterceptor,
			telemetry.ClientLogInterceptor,
		},
	)

	dataCollector := telemetry.NewDataCollector(logger)
	inject.Injector.BindType(new(telemetry.DataCollector), func() interface{} {
		return dataCollector
	})
	appsDI.Injector.BindType(new(telemetry.DataCollector), func() interface{} {
		return dataCollector
	})

	gitCommitLink := fmt.Sprintf("https://github.com/%s/%s/commit/%s",
		cfg.GitRepoOwner,
		cfg.GitRepoName,
		cfg.GitLongCommitHash)
	dataCollector.Logger.Log(telemetry.Info, telemetry.Props{
		telemetry.MessageProp: gitCommitLink,
	})
	err = sqldb.Use(dataCollector, cfg.Config, func(sqlDB *sql.DB) *errs.Error {
		internalErr := sqldb.MigrateUp(dataCollector, sqlDB, "migrations", 0)
		if internalErr != nil {
			dataCollector.Logger.Log(telemetry.Fatal, telemetry.Props{telemetry.CauseProp: internalErr})
			return internalErr
		}

		realTimeStateSyncer := dep.InitRealTimeStateSyncer(dataCollector, sqlDB)
		return startServiceRunner(dataCollector, cfg, sqlDB, realTimeStateSyncer)
	})
	if err != nil {
		dataCollector.Logger.Log(telemetry.Fatal, telemetry.Props{telemetry.CauseProp: err})
		panic(err)
	}
}

func startServiceRunner(
	dataCollector telemetry.DataCollector,
	cfg config.App,
	sqlDB *sql.DB,
	realTimeStateSyncer *realtime.StateSyncer,
) *errs.Error {
	runnerConfig, internalErr := runner.ServiceRunnerConfigFromEnv(dataCollector)
	if internalErr != nil {
		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	githubCfg, internalErr := github.AppConfigFromEnv()
	if internalErr != nil {
		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	prom := metrics.NewPrometheus(appName, serviceName, cfg.Environment)
	nw := network.NewSocket()
	retryFactory := makeRetryFactory(cfg)
	cloudClientRegistry, err := cloudAPI.NewClientRegistry(
		dataCollector,
		nw,
		prom,
		rpc.ConnectionConfig{
			Host:          cfg.CloudGRPCAPIHost,
			Port:          cfg.CloudGRPCAPIPort,
			ShouldEncrypt: cfg.CloudGRPCAPIShouldEncrypt,
			GetAccessToken: func() string {
				return cfg.TeamyServiceAccountAPIToken
			},
			RequestTimeout: cfg.RequestTimeout,
		}, retryFactory)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	teamyClientRegistry, internalErr := api.NewClientRegistry(
		dataCollector,
		nw,
		prom,
		rpc.ConnectionConfig{
			Host:          cfg.TeamyAPIHost,
			Port:          cfg.TeamyAPIPort,
			ShouldEncrypt: cfg.TeamyAPIShouldEncrypt,
			GetAccessToken: func() string {
				return cfg.AppsServiceAccountAPIToken
			},
			RequestTimeout: cfg.RequestTimeout,
		}, retryFactory)
	if internalErr != nil {
		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	privateKeyPEM, err := os.ReadFile(githubCfg.PrivateKeyPEMFilePath)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	githubAppAPI, err := appsDep.InitGithubAppAPI(
		dataCollector,
		cloudClientRegistry,
		teamyClientRegistry,
		http.DefaultClient,
		githubCfg,
		privateKeyPEM,
		sqlDB)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	realTimeStateSyncAPI := dep.InitRealTimeStateSyncAPI(
		dataCollector,
		realTimeStateSyncer)
	graphQLAPI, err := dep.InitGraphQLAPI(
		appName,
		serviceName,
		cfg.Environment,
		dataCollector,
		dep.CloudWebAPIExternalBaseURL(cfg.CloudWebAPIExternalBaseURL),
		cloudClientRegistry,
		realTimeStateSyncer,
		sqlDB)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	taskRPCAPI := dep.InitTaskRPCAPI(
		dataCollector,
		cloudClientRegistry,
		realTimeStateSyncer,
		sqlDB)
	taskLinkRPCAPI := dep.InitTaskLinkRPCAPI(
		dataCollector,
		cloudClientRegistry,
		realTimeStateSyncer,
		sqlDB,
	)
	sprintRPCAPI := dep.InitSprintRPCAPI(
		dataCollector,
		cloudClientRegistry,
		realTimeStateSyncer,
		sqlDB)
	rn := runner.NewServiceRunnerBuilder(
		dataCollector,
		nw,
		prom,
		runnerConfig,
		fullServiceName,
		[]runner.Service{
			githubAppAPI,
			graphQLAPI,
			realTimeStateSyncAPI,
			taskRPCAPI,
			taskLinkRPCAPI,
			sprintRPCAPI,
		}).
		Build()
	rn.Start(nil)
	return nil
}

func getEnv(name string, defaultVal string) string {
	value := os.Getenv(name)
	if len(value) > 0 {
		return value
	}

	return defaultVal
}

func newLineFormatter(environment env.Environment) telemetry.LineFormatter {
	if environment == env.DevelopmentEnv {
		return telemetry.NewOrderedColumnLineFormatter([]string{
			telemetry.HappenAtProp,
			telemetry.SeverityProp,
			telemetry.FileNameProp,
			telemetry.LineNumberProp,
			telemetry.TraceIDProp,
			telemetry.SpanIDProp,
			telemetry.RequestIDProp,
			realtime.MutationIDProp,
			telemetry.ClientIDProp,
			middleware.ProtocolProp,
			middleware.StageProp,
			middleware.HostProp,
			middleware.MethodProp,
			middleware.PathProp,
			middleware.HeadersProp,
			middleware.MetadataProp,
			middleware.BodySizeProp,
			middleware.BodyProp,
			telemetry.CauseProp,
			telemetry.StackTraceProp,
			telemetry.MessageProp,
		})
	}

	return telemetry.NewJSONLineFormatter()
}

func makeRetryFactory(cfg config.App) func() retry.Retry {
	return func() retry.Retry {
		exponentialBackOff := backoff.NewExponentialBuilder().Build()
		return retry.NewMaxCount(runtime.NewBuiltInRuntime(), exponentialBackOff, cfg.RequestRetryMaxCount)
	}
}
