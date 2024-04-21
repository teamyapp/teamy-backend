package cache_test

import (
	"context"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/cache"
	"github.com/teamyapp/teamy-backend/core/cache/cachetest"
)

func BenchmarkTimeBasedCache_Contains(b *testing.B) {
	lineFormatter := telemetry.NewOrderedColumnLineFormatter([]string{})
	logger := telemetry.NewLogger(lineFormatter, os.Stdout, telemetry.Off, []telemetry.LogInterceptor{})
	noopMetrics := cachetest.NewNoopMetrics()
	lruFactory := cache.NewLRUFactory[string, any](logger, noopMetrics, 100)
	timeBasedCache, err := cache.NewTimeBasedCache[string, string](logger, noopMetrics, lruFactory, 100, 2000)
	require.Nil(b, err)

	for i := 0; i < 1000; i++ {
		ct := context.Background()
		num := strconv.FormatInt(int64(i), 10)
		err = timeBasedCache.SetIfExpired(ct, num, num)
		require.Nil(b, err)
	}

	b.ResetTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ct := context.Background()
			num := strconv.FormatInt(int64(i), 10)
			timeBasedCache.Contains(ct, num)
		}(i)
	}

	wg.Wait()
}

func BenchmarkTimeBasedCache_Get(b *testing.B) {
	lineFormatter := telemetry.NewOrderedColumnLineFormatter([]string{})
	logger := telemetry.NewLogger(lineFormatter, os.Stdout, telemetry.Off, []telemetry.LogInterceptor{})
	noopMetrics := cachetest.NewNoopMetrics()

	lruFactory := cache.NewLRUFactory[string, any](logger, noopMetrics, 100)
	timeBasedCache, err := cache.NewTimeBasedCache[string, string](logger, noopMetrics, lruFactory, 100, 2000)
	require.Nil(b, err)

	for i := 0; i < 100; i++ {
		ct := context.Background()
		num := strconv.FormatInt(int64(i), 10)
		err = timeBasedCache.SetIfExpired(ct, num, num)
		require.Nil(b, err)
	}

	b.ResetTimer()

	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ct := context.Background()
			num := strconv.FormatInt(int64(i), 10)
			_, _ = timeBasedCache.Get(ct, num)
		}(i)
	}

	wg.Wait()
}

func BenchmarkTimeBasedCache_SetIfExpired(b *testing.B) {
	lineFormatter := telemetry.NewOrderedColumnLineFormatter([]string{})
	logger := telemetry.NewLogger(lineFormatter, os.Stdout, telemetry.Off, []telemetry.LogInterceptor{})
	noopMetrics := cachetest.NewNoopMetrics()
	lruFactory := cache.NewLRUFactory[string, any](logger, noopMetrics, 100)
	timeBasedCache, err := cache.NewTimeBasedCache[string, string](logger, noopMetrics, lruFactory, 100, 2000)
	require.Nil(b, err)

	b.ResetTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ct := context.Background()
			num := strconv.FormatInt(int64(i), 10)
			err = timeBasedCache.SetIfExpired(ct, num, num)
			require.Nil(b, err)
		}(i)
	}

	wg.Wait()
}

func BenchmarkTimeBasedCache_Remove(b *testing.B) {
	lineFormatter := telemetry.NewOrderedColumnLineFormatter([]string{})
	logger := telemetry.NewLogger(lineFormatter, os.Stdout, telemetry.Off, []telemetry.LogInterceptor{})
	noopMetrics := cachetest.NewNoopMetrics()
	lruFactory := cache.NewLRUFactory[string, any](logger, noopMetrics, 100)
	timeBasedCache, err := cache.NewTimeBasedCache[string, string](logger, noopMetrics, lruFactory, 100, 2000)
	require.Nil(b, err)

	for i := 0; i < 100; i++ {
		ct := context.Background()
		num := strconv.FormatInt(int64(i)%100, 10)
		err = timeBasedCache.SetIfExpired(ct, num, num)
		require.Nil(b, err)
	}

	b.ResetTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ct := context.Background()
			num := strconv.FormatInt(int64(i), 10)
			_ = timeBasedCache.Remove(ct, num)
		}(i)
	}

	wg.Wait()
}
