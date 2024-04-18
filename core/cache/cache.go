package cache

type Cache[Key comparable, Value any] interface {
	Get(key Key) (Value, error)
	Contains(Key Key) bool
	Set(key Key, value Value) error
	Remove(key Key) error
	Size() int
}
