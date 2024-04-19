package cachetest

import "github.com/teamyapp/teamy-backend/core/cache"

type NoopMetrics struct {
}

var _ cache.Metrics = (*NoopMetrics)(nil)

func (n NoopMetrics) ReportCacheHit(key string) {
}

func (n NoopMetrics) ReportCacheMiss(key string) {
}

func NewNoopMetrics() NoopMetrics {
	return NoopMetrics{}
}
