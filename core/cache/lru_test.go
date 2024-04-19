package cache_test

import (
	"github.com/stretchr/testify/require"
	"github.com/teamyapp/teamy-backend/core/cache"
	"github.com/teamyapp/teamy-backend/core/cache/cachetest"
	"testing"
)

type OperationType string

const (
	Get    OperationType = "get"
	Set    OperationType = "set"
	Remove OperationType = "remove"
)

type Operation struct {
	Type  OperationType
	Key   string
	Value string
}

func TestNewLRU(t *testing.T) {
	testCases := []struct {
		name      string
		capacity  int
		expectErr error
	}{
		{
			name:      "valid capacity",
			capacity:  10,
			expectErr: nil,
		},
		{
			name:      "invalid capacity",
			capacity:  0,
			expectErr: cache.InvalidCapacityErr{Capacity: 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := cachetest.NewNoopMetrics()
			_, err := cache.NewLRU[string, string](metrics, tc.capacity)
			require.Equal(t, tc.expectErr, err)
		})
	}
}

func TestLRU_Operations(t *testing.T) {
	testCases := []struct {
		name           string
		capacity       int
		operations     []Operation
		expectedSize   []int
		expectedResult []any
		expectedErrs   []error
	}{
		{
			name:     "set the same key",
			capacity: 3,
			operations: []Operation{
				{Type: Set, Key: "a", Value: "1"},
				{Type: Set, Key: "b", Value: "2"},
				{Type: Set, Key: "c", Value: "3"},
				{Type: Set, Key: "a", Value: "4"},
				{Type: Get, Key: "a"},
				{Type: Get, Key: "b"},
				{Type: Get, Key: "c"},
				{Type: Set, Key: "b", Value: "5"},
				{Type: Get, Key: "b"},
			},
			expectedSize: []int{1, 2, 3, 3, 3, 3, 3, 3, 3},
			expectedResult: []any{
				nil, nil, nil, nil, "4", "2", "3", nil, "5",
			},
			expectedErrs: []error{
				nil, nil, nil, nil, nil, nil, nil, nil, nil,
			},
		},
		{
			name:     "remove key",
			capacity: 3,
			operations: []Operation{
				{Type: Set, Key: "a", Value: "1"},
				{Type: Set, Key: "b", Value: "2"},
				{Type: Set, Key: "c", Value: "3"},
				{Type: Set, Key: "d", Value: "4"},
				{Type: Set, Key: "e", Value: "5"},
				{Type: Remove, Key: "a"},
				{Type: Remove, Key: "b"},
				{Type: Remove, Key: "c"},
				{Type: Get, Key: "a"},
				{Type: Get, Key: "b"},
				{Type: Get, Key: "c"},
				{Type: Get, Key: "d"},
				{Type: Get, Key: "e"},
			},
			expectedSize: []int{1, 2, 3, 3, 3, 3, 3, 2, 2, 2, 2, 2, 2},
			expectedResult: []any{
				nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "4", "5",
			},
			expectedErrs: []error{
				nil, nil, nil, nil, nil,
				cache.KeyNotFoundErr[string]{Key: "a"},
				cache.KeyNotFoundErr[string]{Key: "b"},
				nil,
				cache.KeyNotFoundErr[string]{Key: "a"},
				cache.KeyNotFoundErr[string]{Key: "b"},
				cache.KeyNotFoundErr[string]{Key: "c"},
				nil, nil,
			},
		},
		{
			name:     "evict oldest key",
			capacity: 3,
			operations: []Operation{
				{Type: Set, Key: "a", Value: "1"},
				{Type: Set, Key: "b", Value: "2"},
				{Type: Set, Key: "c", Value: "3"},
				{Type: Set, Key: "d", Value: "4"},
				{Type: Set, Key: "e", Value: "5"},
				{Type: Set, Key: "f", Value: "6"},
				{Type: Set, Key: "g", Value: "7"},
				{Type: Get, Key: "a"},
				{Type: Get, Key: "b"},
				{Type: Get, Key: "c"},
				{Type: Get, Key: "d"},
				{Type: Get, Key: "e"},
				{Type: Get, Key: "f"},
				{Type: Get, Key: "g"},
			},
			expectedSize:   []int{1, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3},
			expectedResult: []any{nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "5", "6", "7"},
			expectedErrs: []error{
				nil, nil, nil, nil, nil, nil, nil,
				cache.KeyNotFoundErr[string]{Key: "a"},
				cache.KeyNotFoundErr[string]{Key: "b"},
				cache.KeyNotFoundErr[string]{Key: "c"},
				cache.KeyNotFoundErr[string]{Key: "d"},
				nil, nil, nil,
			},
		},
		{
			name:     "evict least recently used key",
			capacity: 3,
			operations: []Operation{
				{Type: Set, Key: "a", Value: "1"},
				{Type: Set, Key: "b", Value: "2"},
				{Type: Set, Key: "c", Value: "3"},
				{Type: Set, Key: "d", Value: "4"},
				{Type: Set, Key: "e", Value: "5"},
				{Type: Set, Key: "f", Value: "6"},
				{Type: Set, Key: "g", Value: "7"},
				{Type: Set, Key: "h", Value: "8"},
				{Type: Get, Key: "g"},
				{Type: Get, Key: "f"},
				{Type: Set, Key: "i", Value: "9"},
				{Type: Get, Key: "h"},
				{Type: Set, Key: "g", Value: "10"},
				{Type: Set, Key: "j", Value: "11"},
				{Type: Get, Key: "f"},
			},
			expectedSize:   []int{1, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3},
			expectedResult: []any{nil, nil, nil, nil, nil, nil, nil, nil, "7", "6", nil, nil, nil, nil, nil},
			expectedErrs: []error{
				nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
				cache.KeyNotFoundErr[string]{Key: "h"},
				nil, nil,
				cache.KeyNotFoundErr[string]{Key: "f"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics := cachetest.NewNoopMetrics()
			lru, err := cache.NewLRU[string, string](metrics, tc.capacity)
			require.Nil(t, err)

			for i, op := range tc.operations {
				switch op.Type {
				case Get:
					value, err := lru.Get(op.Key)
					require.Equal(t, tc.expectedErrs[i], err)
					if err != nil {
						continue
					}

					require.Equal(t, tc.expectedResult[i], value)
				case Set:
					err = lru.Set(op.Key, op.Value)
					require.Equal(t, tc.expectedErrs[i], err)
					if err != nil {
						continue
					}
				case Remove:
					err = lru.Remove(op.Key)
					require.Equal(t, tc.expectedErrs[i], err)
					if err != nil {
						continue
					}
				}

				require.Equal(t, tc.expectedSize[i], lru.Size())
			}
		})
	}
}

func BenchmarkLRU_Contains(b *testing.B) {
	metrics := cachetest.NewNoopMetrics()
	lru, err := cache.NewLRU[string, string](metrics, 1000)
	require.Nil(b, err)

	for i := 0; i < 1000; i++ {
		err = lru.Set(string(rune(i)), string(rune(i)))
		require.Nil(b, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lru.Contains("a")
	}
}

func BenchmarkLRU_Get(b *testing.B) {
	metrics := cachetest.NewNoopMetrics()
	lru, err := cache.NewLRU[string, string](metrics, 100)
	require.Nil(b, err)

	for i := 0; i < 100; i++ {
		err = lru.Set(string(rune(i)), string(rune(i)))
		require.Nil(b, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = lru.Get("a")
	}
}

func BenchmarkLRU_Set(b *testing.B) {
	metrics := cachetest.NewNoopMetrics()
	lru, err := cache.NewLRU[string, string](metrics, 100)
	require.Nil(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = lru.Set(string(rune(i)), string(rune(i)))
		require.Nil(b, err)
	}
}

func BenchmarkLRU_Remove(b *testing.B) {
	metrics := cachetest.NewNoopMetrics()
	lru, err := cache.NewLRU[string, string](metrics, 100)
	require.Nil(b, err)

	for i := 0; i < 100; i++ {
		err = lru.Set(string(rune(i)), string(rune(i)))
		require.Nil(b, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = lru.Remove("a")
	}
}
