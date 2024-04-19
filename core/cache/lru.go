package cache

type KeyValuePair[Key comparable, Value any] struct {
	Key   Key
	Value Value
}

type LRU[Key comparable, Value any] struct {
	capacity   int
	size       int
	index      map[Key]*Buffer[*KeyValuePair[Key, Value]]
	bufferHead *Buffer[*KeyValuePair[Key, Value]]
	bufferTail *Buffer[*KeyValuePair[Key, Value]]
}

var _ Cache[string, int] = (*LRU[string, int])(nil)

func (l *LRU[Key, Value]) Get(key Key) (Value, error) {
	buffer, ok := l.index[key]
	if !ok {
		return *(new(Value)), KeyNotFoundErr[Key]{Key: key}
	}

	l.removeBuffer(buffer)
	l.addBufferToTheTail(buffer)
	return buffer.data.Value, nil
}

func (l *LRU[Key, Value]) Contains(key Key) bool {
	_, ok := l.index[key]
	return ok
}

func (l *LRU[Key, Value]) Set(key Key, value Value) error {
	buffer, ok := l.index[key]
	if ok {
		l.removeBuffer(buffer)
		buffer.data.Value = value
		l.addBufferToTheTail(buffer)
		return nil
	}

	if l.size == l.capacity {
		l.evict()
	}

	buffer = &Buffer[*KeyValuePair[Key, Value]]{
		data: &KeyValuePair[Key, Value]{
			Key:   key,
			Value: value,
		},
	}
	l.index[key] = buffer
	l.size++
	l.addBufferToTheTail(buffer)
	return nil
}

func (l *LRU[Key, Value]) Remove(key Key) error {
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
	return l.size
}

func (l *LRU[Key, Value]) Keys() []Key {
	keys := make([]Key, l.size)
	buffer := l.bufferHead
	for i := 0; buffer != nil; i++ {
		keys[i] = buffer.data.Key
		buffer = buffer.next
	}

	return keys
}

func (l *LRU[Key, Value]) Entries() []KeyValuePair[Key, Value] {
	entries := make([]KeyValuePair[Key, Value], l.size)
	buffer := l.bufferHead
	for i := 0; buffer != nil; i++ {
		entries[i] = KeyValuePair[Key, Value]{
			Key:   buffer.data.Key,
			Value: buffer.data.Value,
		}
		buffer = buffer.next
	}

	return entries
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
	delete(l.index, l.bufferHead.data.Key)
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
