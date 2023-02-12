package main

import (
	"database/sql"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/cloud/libs/env"
	"github.com/teamyapp/cloud/libs/errs"
	tmio "github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/cloud/libs/middleware"
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

var serviceLabels = []string{"teamy", "backend"}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	cfg, err := config.AppFromEnv()
	if err != nil {
		panic(err)
	}

	lineFormatter := newLineFormatter(cfg.Environment)
	logOutput, err := newLogOutput(cfg.Environment, strings.Join(serviceLabels, "-"))
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
		if err != nil {
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

	exponentialBackOff := backoff.NewExponentialBuilder().Build()
	maxCountRetry := retry.NewMaxCount(runtime.NewBuiltInRuntime(), exponentialBackOff, cfg.RequestRetryMaxCount)
	cloudClientRegistry, err := cloudAPI.NewClientRegistry(
		dataCollector,
		rpc.ConnectionConfig{
			Host:          cfg.CloudGRPCAPIHost,
			Port:          cfg.CloudGRPCAPIPort,
			ShouldEncrypt: cfg.CloudGRPCAPIShouldEncrypt,
			GetAccessToken: func() string {
				return cfg.TeamyServiceAccountAPIToken
			},
			RequestTimeout: cfg.RequestTimeout,
		}, maxCountRetry)
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
		rpc.ConnectionConfig{
			Host:          cfg.TeamyAPIHost,
			Port:          cfg.TeamyAPIPort,
			ShouldEncrypt: cfg.TeamyAPIShouldEncrypt,
			GetAccessToken: func() string {
				return cfg.AppsServiceAccountAPIToken
			},
			RequestTimeout: cfg.RequestTimeout,
		}, maxCountRetry)
	if internalErr != nil {
		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	githubAppAPI, err := appsDep.InitGithubAppAPI(
		dataCollector,
		cloudClientRegistry,
		teamyClientRegistry,
		githubCfg,
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
		runnerConfig, []runner.Service{
			githubAppAPI,
			graphQLAPI,
			realTimeStateSyncAPI,
			taskRPCAPI,
			taskLinkRPCAPI,
			sprintRPCAPI,
		}).
		Build()
	rn.Start()
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

func newLogOutput(environment env.Environment, serviceName string) (io.WriteCloser, *errs.Error) {
	if environment == env.DevelopmentEnv {
		logFileName := fmt.Sprintf("%v.log", serviceName)
		logFilePath := getEnv("LOG_OUTPUT_FILE", filepath.Join("..", "logs", logFileName))
		logDir := filepath.Dir(logFilePath)

		// MkdirAll requires at least 700 permission:
		// https://github.com/golang/go/issues/22323
		err := os.MkdirAll(logDir, 0744)
		if err != nil {
			return nil, &errs.Error{
				Code:     errs.OS,
				EmbedErr: err,
			}
		}

		file, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640)
		if err != nil {
			return nil, &errs.Error{
				Code:     errs.IO,
				EmbedErr: err,
			}
		}

		return tmio.NewMultiWriteCloser(file, os.Stdout), nil
	}

	return os.Stdout, nil
}
