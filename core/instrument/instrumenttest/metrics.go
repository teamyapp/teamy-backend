package instrumenttest

import (
	"time"

	"github.com/teamyapp/teamy-backend/core/cache"

	"github.com/teamyapp/cloud/libs/middleware"
	"github.com/teamyapp/teamy-backend/core/transaction"
)

type NoopMetrics struct {
}

var _ middleware.ClientHTTPMetrics = (*NoopMetrics)(nil)
var _ middleware.ServerHTTPMetrics = (*NoopMetrics)(nil)
var _ middleware.ClientGRPCMetrics = (*NoopMetrics)(nil)
var _ middleware.ServerGRPCMetrics = (*NoopMetrics)(nil)
var _ transaction.Metrics = (*NoopMetrics)(nil)
var _ cache.Metrics = (*NoopMetrics)(nil)

func (n NoopMetrics) ReportHTTPIncomingRequest(method string, pattern string) {
}

func (n NoopMetrics) ReportHTTPIncomingRequestResponseTime(method string, pattern string, duration time.Duration) {
}

func (n NoopMetrics) ReportHTTPOutgoingRequest(target string, method string, pattern string) {
}

func (n NoopMetrics) ReportHTTPOutgoingRequestResponseTime(target string, method string, pattern string, duration time.Duration) {
}

func (n NoopMetrics) ReportGRPCIncomingRequest(method string) {
}

func (n NoopMetrics) ReportGRPCIncomingRequestResponseTime(method string, duration time.Duration) {
}

func (n NoopMetrics) ReportGRPCOutgoingRequest(target string, method string) {
}

func (n NoopMetrics) ReportGRPCOutgoingRequestResponseTime(target string, method string, duration time.Duration) {
}

func (n NoopMetrics) ReportTransactionBegin() {
}

func (n NoopMetrics) ReportTransactionCommit() {
}

func (n NoopMetrics) ReportTransactionRollback() {
}

func (n NoopMetrics) ReportTransactionDuration(duration time.Duration, result transaction.Result) {
}

func (n NoopMetrics) ReportCacheHit(cacheName string) {
}

func (n NoopMetrics) ReportCacheMiss(cacheName string) {
}

func (n NoopMetrics) ReportCacheEviction(cacheName string) {
}

func (n NoopMetrics) ReportCacheHitDuration(cacheName string, duration time.Duration) {
}

func (n NoopMetrics) ReportCacheMissDuration(cacheName string, duration time.Duration) {
}

func NewNoopMetrics() NoopMetrics {
	return NoopMetrics{}
}
