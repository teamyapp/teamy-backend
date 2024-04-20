package cache

import "context"

type Cache[Key comparable, Value any] interface {
	Get(ct context.Context, key Key) (Value, error)
	Contains(ct context.Context, Key Key) bool
	Set(ct context.Context, key Key, value Value) error
	Remove(ct context.Context, key Key) error
	Size(ct context.Context) int
	Keys(ct context.Context) []Key
	Entries(ct context.Context) []KeyValuePair[Key, Value]
}

type Factory[Key comparable, Value any] interface {
	MakeCache() (Cache[Key, Value], error)
}
