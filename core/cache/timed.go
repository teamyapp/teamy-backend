package cache

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/teamyapp/cloud/libs/telemetry"
)

const timeBasedCacheName = "TimeBasedCache"

type ValueWithExpiration[Value any] struct {
	value    Value
	expireAt time.Time
}

type Bucket[Key comparable, Value any] struct {
	mu    sync.Mutex
	cache Cache[Key, Value]
}

type TimeBasedCache[Key comparable, Value any] struct {
	logger  telemetry.Logger
	metrics Metrics
	ttl     time.Duration
	buckets []*Bucket[Key, any]
}

func (t *TimeBasedCache[Key, Value]) Get(ct context.Context, key Key) (Value, error) {
	now := time.Now()
	bucket := t.getBucket(key)
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	v, err := bucket.cache.Get(ct, key)
	if err != nil {
		var keyNotFoundErr KeyNotFoundErr[Key]
		if errors.As(err, &keyNotFoundErr) {
			t.metrics.ReportCacheMiss(timeBasedCacheName)
			t.metrics.ReportCacheMissDuration(timeBasedCacheName, time.Now().Sub(now))
			t.logger.DebugWithContext(ct, fmt.Sprintf("time based cache miss: key=%v", key))
		}

		return *new(Value), err
	}

	valueWithExpiration := v.(ValueWithExpiration[Value])
	if valueWithExpiration.expireAt.Before(now) {
		err = bucket.cache.Remove(ct, key)
		if err != nil {
			return *new(Value), err
		}

		t.metrics.ReportCacheMiss(timeBasedCacheName)
		t.metrics.ReportCacheMissDuration(timeBasedCacheName, time.Now().Sub(now))
		t.logger.DebugWithContext(ct, fmt.Sprintf("time based cache miss: key=%v", key))
		return *new(Value), KeyNotFoundErr[Key]{Key: key}
	}

	t.metrics.ReportCacheHit(timeBasedCacheName)
	t.metrics.ReportCacheHitDuration(timeBasedCacheName, time.Now().Sub(now))
	t.logger.DebugWithContext(ct, fmt.Sprintf("time based cache hit: key=%v", key))
	return valueWithExpiration.value, nil
}

func (t *TimeBasedCache[Key, Value]) Contains(ct context.Context, key Key) bool {
	bucket := t.getBucket(key)
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	v, err := bucket.cache.Get(ct, key)
	if err != nil {
		return false
	}

	valueWithExpiration := v.(ValueWithExpiration[Value])
	return !now.After(valueWithExpiration.expireAt)
}

func (t *TimeBasedCache[Key, Value]) SetIfExpired(ct context.Context, key Key, value Value) error {
	bucket := t.getBucket(key)
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	v, err := bucket.cache.Get(ct, key)
	if err == nil {
		valueWithExpiration := v.(ValueWithExpiration[Value])
		if !now.After(valueWithExpiration.expireAt) {
			return nil
		}
	}

	var keyNotFoundErr KeyNotFoundErr[Key]
	if !errors.As(err, &keyNotFoundErr) {
		return err
	}

	return bucket.cache.Set(ct, key, ValueWithExpiration[Value]{
		value:    value,
		expireAt: now.Add(t.ttl),
	})
}

func (t *TimeBasedCache[Key, Value]) Remove(ct context.Context, key Key) error {
	bucket := t.getBucket(key)
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now()
	v, err := bucket.cache.Get(ct, key)
	if err != nil {
		return err
	}

	valueWithExpiration := v.(ValueWithExpiration[Value])
	if now.After(valueWithExpiration.expireAt) {
		err = bucket.cache.Remove(ct, key)
		if err != nil {
			return err
		}

		return KeyNotFoundErr[Key]{Key: key}
	}

	return bucket.cache.Remove(ct, key)
}

func (t *TimeBasedCache[Key, Value]) Size(ct context.Context) int {
	for _, bucket := range t.buckets {
		bucket.mu.Lock()
	}

	defer func() {
		for _, bucket := range t.buckets {
			bucket.mu.Unlock()
		}
	}()

	count := 0
	now := time.Now()

	for _, bucket := range t.buckets {
		for _, entry := range bucket.cache.Entries(ct) {
			valueWithExpiration := entry.Value.(ValueWithExpiration[Value])
			if !now.After(valueWithExpiration.expireAt) {
				count++
			}
		}
	}

	return count
}

func (t *TimeBasedCache[Key, Value]) Keys(ct context.Context) []Key {
	for _, bucket := range t.buckets {
		bucket.mu.Lock()
	}

	defer func() {
		for _, bucket := range t.buckets {
			bucket.mu.Unlock()
		}
	}()

	keys := make([]Key, 0)
	now := time.Now()
	for _, bucket := range t.buckets {
		for _, entry := range bucket.cache.Entries(ct) {
			valueWithExpiration := entry.Value.(ValueWithExpiration[Value])
			if !now.After(valueWithExpiration.expireAt) {
				keys = append(keys, entry.Key)
			}
		}
	}

	return keys
}

func (t *TimeBasedCache[Key, Value]) Entries(ct context.Context) []KeyValuePair[Key, Value] {
	for _, bucket := range t.buckets {
		bucket.mu.Lock()
	}

	defer func() {
		for _, bucket := range t.buckets {
			bucket.mu.Unlock()
		}
	}()

	entries := make([]KeyValuePair[Key, Value], 0)
	now := time.Now()
	for _, bucket := range t.buckets {
		for _, entry := range bucket.cache.Entries(ct) {
			valueWithExpiration := entry.Value.(ValueWithExpiration[Value])
			if !now.After(valueWithExpiration.expireAt) {
				entries = append(entries, KeyValuePair[Key, Value]{
					Key:   entry.Key,
					Value: valueWithExpiration.value,
				})
			}
		}
	}

	return entries
}

func (t *TimeBasedCache[Key, Value]) getBucket(key Key) *Bucket[Key, any] {
	return t.buckets[t.getBucketIndex(key)]
}

func (t *TimeBasedCache[Key, Value]) getBucketIndex(key Key) int {
	keyStr := fmt.Sprintf("%v", key)
	h := fnv.New32a()
	_, _ = h.Write([]byte(keyStr))
	hashSum := h.Sum32()
	return int(hashSum) % len(t.buckets)
}

func NewTimeBasedCache[Key comparable, Value any](
	logger telemetry.Logger,
	metrics Metrics,
	cacheFactory Factory[Key, any],
	bucketCount int,
	ttl time.Duration,
) (*TimeBasedCache[Key, Value], error) {
	buckets := make([]*Bucket[Key, any], bucketCount)
	for i := 0; i < bucketCount; i++ {
		cache, err := cacheFactory.MakeCache()
		if err != nil {
			return nil, err
		}

		buckets[i] = &Bucket[Key, any]{
			cache: cache,
		}
	}

	return &TimeBasedCache[Key, Value]{
		logger:  logger,
		metrics: metrics,
		ttl:     ttl,
		buckets: buckets,
	}, nil
}
