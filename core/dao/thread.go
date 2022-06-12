package dao

type Thread interface {
	CreateThread(threadID uint64) error
	DeleteThread(threadID uint64) error
}
