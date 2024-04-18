package cache

type Buffer[Data any] struct {
	data Data
	prev *Buffer[Data]
	next *Buffer[Data]
}
