package instrument

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/teamyapp/cloud/libs/env"
	"github.com/teamyapp/cloud/libs/metrics"
	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/teamy-backend/core/cache"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/transaction"
)

type Prometheus struct {
	metrics.Prometheus
	transactionBeginCountMetric    *prometheus.CounterVec
	transactionCommitCountMetric   *prometheus.CounterVec
	transactionRollbackCountMetric *prometheus.CounterVec
	transactionDurationMetric      *prometheus.HistogramVec
	daoOperationMetric             *prometheus.CounterVec
	cacheActivityCountMetric       *prometheus.CounterVec
	cacheActivityDurationMetric    *prometheus.HistogramVec
}

var _ middleware.ClientHTTPMetrics = (*Prometheus)(nil)
var _ middleware.ServerHTTPMetrics = (*Prometheus)(nil)
var _ middleware.ClientGRPCMetrics = (*Prometheus)(nil)
var _ middleware.ServerGRPCMetrics = (*Prometheus)(nil)
var _ transaction.Metrics = (*Prometheus)(nil)
var _ dao.Metrics = (*Prometheus)(nil)
var _ cache.Metrics = (*Prometheus)(nil)

func (p Prometheus) ReportTransactionBegin() {
	p.transactionBeginCountMetric.WithLabelValues().Inc()
}

func (p Prometheus) ReportTransactionCommit() {
	p.transactionCommitCountMetric.WithLabelValues().Inc()
}

func (p Prometheus) ReportTransactionRollback() {
	p.transactionRollbackCountMetric.WithLabelValues().Inc()
}

func (p Prometheus) ReportTransactionDuration(duration time.Duration, result transaction.Result) {
	p.transactionDurationMetric.WithLabelValues(string(result)).Observe(float64(duration.Milliseconds()))
}

func (p Prometheus) ReportDaoOperation(daoName string, operation string) {
	p.daoOperationMetric.WithLabelValues(daoName, operation).Inc()
}

func (p Prometheus) ReportCacheHit(cacheName string) {
	p.cacheActivityCountMetric.WithLabelValues(cacheName, "hit").Inc()
}

func (p Prometheus) ReportCacheMiss(cacheName string) {
	p.cacheActivityCountMetric.WithLabelValues(cacheName, "miss").Inc()
}

func (p Prometheus) ReportCacheEviction(cacheName string) {
	p.cacheActivityCountMetric.WithLabelValues(cacheName, "eviction").Inc()
}

func (p Prometheus) ReportCacheHitDuration(cacheName string, duration time.Duration) {
	p.cacheActivityDurationMetric.WithLabelValues(cacheName, "hit").Observe(float64(duration.Microseconds()))
}

func (p Prometheus) ReportCacheMissDuration(cacheName string, duration time.Duration) {
	p.cacheActivityDurationMetric.WithLabelValues(cacheName, "miss").Observe(float64(duration.Microseconds()))
}

func NewPrometheus(appMame string, serviceName string, environment env.Environment) Prometheus {
	transactionBeginCountMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "transaction_begin_count",
			ConstLabels: map[string]string{
				metrics.EnvironmentLabel: string(environment),
			},
		},
		[]string{})
	transactionCommitCountMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "transaction_commit_count",
			ConstLabels: map[string]string{
				metrics.EnvironmentLabel: string(environment),
			},
		},
		[]string{})
	transactionRollbackCountMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "transaction_rollback_count",
			ConstLabels: map[string]string{
				metrics.EnvironmentLabel: string(environment),
			},
		},
		[]string{})
	transactionDurationMetric := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "transaction_duration",
			ConstLabels: map[string]string{
				metrics.EnvironmentLabel: string(environment),
			},
			Buckets: prometheus.ExponentialBuckets(1, 2, 20),
		},
		[]string{
			"result",
		})
	daoOperationMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "dao_operation_count",
			ConstLabels: map[string]string{
				metrics.EnvironmentLabel: string(environment),
			},
		},
		[]string{
			"dao",
			"operation",
		})
	cacheActivityCountMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "cache_activity_count",
			ConstLabels: map[string]string{
				metrics.EnvironmentLabel: string(environment),
			},
		},
		[]string{
			"cacheName",
			"activity",
		})
	cacheActivityDurationMetric := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: appMame,
			Subsystem: serviceName,
			Name:      "cache_activity_duration",
			ConstLabels: map[string]string{
				metrics.EnvironmentLabel: string(environment),
			},
			Buckets: prometheus.ExponentialBuckets(1, 2, 30),
		},
		[]string{
			"cacheName",
			"activity",
		})
	prometheus.MustRegister(transactionBeginCountMetric)
	prometheus.MustRegister(transactionCommitCountMetric)
	prometheus.MustRegister(transactionRollbackCountMetric)
	prometheus.MustRegister(transactionDurationMetric)
	prometheus.MustRegister(daoOperationMetric)
	prometheus.MustRegister(cacheActivityCountMetric)
	prometheus.MustRegister(cacheActivityDurationMetric)
	return Prometheus{
		Prometheus:                     metrics.NewPrometheus(appMame, serviceName, environment),
		transactionBeginCountMetric:    transactionBeginCountMetric,
		transactionCommitCountMetric:   transactionCommitCountMetric,
		transactionRollbackCountMetric: transactionRollbackCountMetric,
		transactionDurationMetric:      transactionDurationMetric,
		daoOperationMetric:             daoOperationMetric,
		cacheActivityCountMetric:       cacheActivityCountMetric,
		cacheActivityDurationMetric:    cacheActivityDurationMetric,
	}
}
