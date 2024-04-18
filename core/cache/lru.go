package cache

import "sync"

type KeyValuePair[Key comparable, Value any] struct {
	key   Key
	value Value
}

type LRU[Key comparable, Value any] struct {
	capacity   int
	size       int
	index      map[Key]*Buffer[*KeyValuePair[Key, Value]]
	bufferHead *Buffer[*KeyValuePair[Key, Value]]
	bufferTail *Buffer[*KeyValuePair[Key, Value]]
	mu         sync.RWMutex
}

var _ Cache[string, int] = (*LRU[string, int])(nil)

func (l *LRU[Key, Value]) Get(key Key) (Value, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	buffer, ok := l.index[key]
	if !ok {
		return *(new(Value)), KeyNotFoundErr[Key]{Key: key}
	}

	l.removeBuffer(buffer)
	l.addBufferToTheTail(buffer)
	return buffer.data.value, nil
}

func (l *LRU[Key, Value]) Contains(key Key) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	_, ok := l.index[key]
	return ok
}

func (l *LRU[Key, Value]) Set(key Key, value Value) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	buffer, ok := l.index[key]
	if ok {
		l.removeBuffer(buffer)
		buffer.data.value = value
		l.addBufferToTheTail(buffer)
		return nil
	}

	if l.size == l.capacity {
		l.evict()
	}

	buffer = &Buffer[*KeyValuePair[Key, Value]]{
		data: &KeyValuePair[Key, Value]{
			key:   key,
			value: value,
		},
	}
	l.index[key] = buffer
	l.size++
	l.addBufferToTheTail(buffer)
	return nil
}

func (l *LRU[Key, Value]) Remove(key Key) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	buffer, ok := l.index[key]
	if !ok {
		return KeyNotFoundErr[Key]{Key: key}
	}

	l.removeBuffer(buffer)
	delete(l.index, key)
	l.size--
	return nil
}

func (l *LRU[Key, Value]) Size() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.size
}

func (l *LRU[Key, Value]) removeBuffer(buffer *Buffer[*KeyValuePair[Key, Value]]) {
	if buffer == l.bufferTail {
		l.bufferTail = buffer.prev
	} else {
		buffer.next.prev = buffer.prev
	}

	if buffer == l.bufferHead {
		l.bufferHead = buffer.next
	} else {
		buffer.prev.next = buffer.next
	}
}

func (l *LRU[Key, Value]) addBufferToTheTail(buffer *Buffer[*KeyValuePair[Key, Value]]) {
	if l.bufferTail == nil {
		l.bufferTail = buffer
		l.bufferHead = buffer
		return
	}

	buffer.prev = l.bufferTail
	l.bufferTail.next = buffer
	l.bufferTail = buffer
}

func (l *LRU[Key, Value]) evict() {
	delete(l.index, l.bufferHead.data.key)
	l.bufferHead = l.bufferHead.next
	if l.bufferHead != nil {
		l.bufferHead.prev = nil
	}

	l.size--
}

func NewLRU[Key comparable, Value any](capacity int) (*LRU[Key, Value], error) {
	if capacity <= 1 {
		return nil, InvalidCapacityErr{Capacity: capacity}
	}

	return &LRU[Key, Value]{
		capacity: capacity,
		size:     0,
		index:    make(map[Key]*Buffer[*KeyValuePair[Key, Value]]),
	}, nil
}
