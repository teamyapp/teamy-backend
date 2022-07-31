package entity

type Iterator[Item any] interface {
	HasNext() (bool, error)
	Next() (Item, error)
}
