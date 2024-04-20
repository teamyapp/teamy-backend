package cache

import "time"

type Metrics interface {
	ReportCacheHit(cacheName string)
	ReportCacheMiss(cacheName string)
	ReportCacheEviction(cacheName string)
	ReportCacheHitDuration(cacheName string, duration time.Duration)
	ReportCacheMissDuration(cacheName string, duration time.Duration)
}
