package cachetest

import (
	"time"

	"github.com/teamyapp/teamy-backend/core/cache"
)

type NoopMetrics struct {
}

var _ cache.Metrics = (*NoopMetrics)(nil)

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
