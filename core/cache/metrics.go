package cache

type Metrics interface {
	ReportCacheHit(key string)
	ReportCacheMiss(key string)
}
