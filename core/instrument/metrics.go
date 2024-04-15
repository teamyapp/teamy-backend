package instrument

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/teamyapp/cloud/libs/env"
	"github.com/teamyapp/cloud/libs/metrics"
	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/teamy-backend/core/transaction"
)

var durationBuckets = []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}

type Prometheus struct {
	metrics.Prometheus
	transactionBeginCountMetric    *prometheus.CounterVec
	transactionCommitCountMetric   *prometheus.CounterVec
	transactionRollbackCountMetric *prometheus.CounterVec
	transactionDurationMetric      *prometheus.HistogramVec
}

var _ middleware.ClientHTTPMetrics = (*Prometheus)(nil)
var _ middleware.ServerHTTPMetrics = (*Prometheus)(nil)
var _ middleware.ClientGRPCMetrics = (*Prometheus)(nil)
var _ middleware.ServerGRPCMetrics = (*Prometheus)(nil)
var _ transaction.Metrics = (*Prometheus)(nil)

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
			Buckets: durationBuckets,
		},
		[]string{
			"result",
		})
	prometheus.MustRegister(transactionBeginCountMetric)
	prometheus.MustRegister(transactionCommitCountMetric)
	prometheus.MustRegister(transactionRollbackCountMetric)
	prometheus.MustRegister(transactionDurationMetric)
	return Prometheus{
		Prometheus:                     metrics.NewPrometheus(appMame, serviceName, environment),
		transactionBeginCountMetric:    transactionBeginCountMetric,
		transactionCommitCountMetric:   transactionCommitCountMetric,
		transactionRollbackCountMetric: transactionRollbackCountMetric,
		transactionDurationMetric:      transactionDurationMetric,
	}
}
