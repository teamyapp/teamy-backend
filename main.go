package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	cloudProto "github.com/teamyapp/cloud/app/api/proto"
	cloudClient "github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/cloud/libs/env"
	"github.com/teamyapp/cloud/libs/errs"
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
	"github.com/teamyapp/teamy-backend/core"
	teamyClient "github.com/teamyapp/teamy-backend/core/client"
	"github.com/teamyapp/teamy-backend/core/dep"
	"github.com/teamyapp/teamy-backend/core/inject"
	"github.com/teamyapp/teamy-backend/core/instrument"
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
	inject.Injector.BindType(new(telemetry.Logger), func() interface{} {
		return logger
	})
	appsDI.Injector.BindType(new(telemetry.Logger), func() interface{} {
		return logger
	})

	gitCommitLink := fmt.Sprintf("https://github.com/%s/%s/commit/%s",
		cfg.GitRepoOwner,
		cfg.GitRepoName,
		cfg.GitLongCommitHash)
	logger.Info(gitCommitLink)
	prom := instrument.NewPrometheus(appName, serviceName, cfg.Environment)
	err = sqldb.Use(logger, cfg.Config, func(sqlDB *sql.DB) *errs.Error {
		internalErr := sqldb.MigrateUp(logger, sqlDB, "migrations", 0)
		if internalErr != nil {
			return internalErr
		}

		realTimeStateSyncer := dep.InitRealTimeStateSyncer(logger, prom, sqlDB)
		return startServiceRunner(logger, prom, cfg, sqlDB, realTimeStateSyncer)
	})
	if err != nil {
		logger.Error(err)
		panic(err)
	}
}

func startServiceRunner(
	logger telemetry.Logger,
	prom instrument.Prometheus,
	cfg config.App,
	sqlDB *sql.DB,
	realTimeStateSyncer *realtime.StateSyncer,
) *errs.Error {
	runnerConfig, internalErr := runner.ServiceRunnerConfigFromEnv()
	if internalErr != nil {
		return internalErr
	}

	githubCfg, internalErr := github.AppConfigFromEnv()
	if internalErr != nil {
		return internalErr
	}

	nw := network.NewSocket()
	retryFactory := makeRetryFactory(logger, cfg)
	cloudClientRegistry, err := cloudClient.NewRegistry(
		logger,
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
		return errs.NewError(errs.Unknown, err.Error())
	}

	authorizationClient := cloudClientRegistry.AuthorizationClient()
	applyAuthorizationCfgReq := &cloudProto.ApplyAuthorizationConfigRequest{
		ConfigContent: core.AuthorizationConfig,
	}
	ct := context.Background()
	ensureSucceed(ct, logger, func() *errs.Error {
		logger.Info("Start applying authorization config")
		_, err = authorizationClient.ApplyAuthorizationConfig(ct, applyAuthorizationCfgReq)
		if err != nil {
			logger.Warning("failed to apply authorization config to cloud")
			return errs.FromGRPCErr(err)
		}

		logger.Info("successfully applied authorization config to cloud")
		return nil
	})
	teamyClientRegistry, internalErr := teamyClient.NewRegistry(
		logger,
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
		return internalErr
	}

	privateKeyPEM, err := os.ReadFile(githubCfg.PrivateKeyPEMFilePath)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	githubAppAPI, err := appsDep.InitGithubAppAPI(
		logger,
		cloudClientRegistry,
		teamyClientRegistry,
		http.DefaultClient, //TODO: add a log middleware to log outgoing request
		githubCfg,
		privateKeyPEM,
		sqlDB,
		appsDep.TeamyWebUIBaseURL(cfg.TeamyWebUIBaseURL),
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	realTimeStateSyncAPI := dep.InitRealTimeStateSyncAPI(
		logger,
		realTimeStateSyncer)
	mapServerURL := dep.MapServerURL(fmt.Sprintf("%s/stream", cfg.CloudWebAPIExternalBaseURL))
	graphQLAPI, err := dep.InitGraphQLAPI(
		appName,
		serviceName,
		cfg.Environment,
		logger,
		prom,
		dep.CloudWebAPIExternalBaseURL(cfg.CloudWebAPIExternalBaseURL),
		mapServerURL,
		cloudClientRegistry,
		realTimeStateSyncer,
		dep.JWTSigningKey(cfg.JWTSigningKey),
		dep.CacheCapacity(cfg.CacheCapacity),
		dep.TimeBasedCacheBucketCount(cfg.TimeBasedCacheBucketCount),
		dep.TimeBasedCacheTTL(cfg.TimeBasedCacheTTL),
		sqlDB)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	taskRPCAPI, err := dep.InitTaskRPCAPI(
		logger,
		prom,
		cloudClientRegistry,
		realTimeStateSyncer,
		dep.CacheCapacity(cfg.CacheCapacity),
		dep.TimeBasedCacheBucketCount(cfg.TimeBasedCacheBucketCount),
		dep.TimeBasedCacheTTL(cfg.TimeBasedCacheTTL),
		sqlDB,
		dep.CloudWebAPIExternalBaseURL(cfg.CloudWebAPIExternalBaseURL))
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	taskLinkRPCAPI, err := dep.InitTaskLinkRPCAPI(
		logger,
		prom,
		cloudClientRegistry,
		realTimeStateSyncer,
		dep.CacheCapacity(cfg.CacheCapacity),
		dep.TimeBasedCacheBucketCount(cfg.TimeBasedCacheBucketCount),
		dep.TimeBasedCacheTTL(cfg.TimeBasedCacheTTL),
		sqlDB,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	sprintRPCAPI, err := dep.InitSprintRPCAPI(
		logger,
		prom,
		cloudClientRegistry,
		realTimeStateSyncer,
		dep.CacheCapacity(cfg.CacheCapacity),
		dep.TimeBasedCacheBucketCount(cfg.TimeBasedCacheBucketCount),
		dep.TimeBasedCacheTTL(cfg.TimeBasedCacheTTL),
		sqlDB)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	teamRPCAPI, err := dep.InitTeamRPCAPI(
		logger,
		prom,
		cloudClientRegistry,
		realTimeStateSyncer,
		dep.CacheCapacity(cfg.CacheCapacity),
		dep.TimeBasedCacheBucketCount(cfg.TimeBasedCacheBucketCount),
		dep.TimeBasedCacheTTL(cfg.TimeBasedCacheTTL),
		sqlDB,
		dep.CloudWebAPIExternalBaseURL(cfg.CloudWebAPIExternalBaseURL))
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	rn := runner.NewServiceRunnerBuilder(
		logger,
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
			teamRPCAPI,
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

func ensureSucceed(ct context.Context, logger telemetry.Logger, execute func() *errs.Error) {
	backOff := backoff.
		NewUniformBuilder().
		Delay(5 * time.Second).
		Build()
	runTime := runtime.NewBuiltInRuntime()
	rt := retry.NewInfinite(logger, runTime, backOff, backOff, nil)
	rt.WithRetry(ct, func() *errs.Error {
		return execute()
	})
}

func makeRetryFactory(logger telemetry.Logger, cfg config.App) func() retry.Retry {
	return func() retry.Retry {
		shortBackOff := backoff.NewExponentialBuilder().Build()
		longBackOff := backoff.NewExponentialBuilder().Build()
		return retry.NewMaxCount(
			logger,
			runtime.NewBuiltInRuntime(),
			shortBackOff,
			longBackOff,
			cfg.RequestRetryMaxCount,
			nil)
	}
}
