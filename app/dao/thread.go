package dao

type Thread interface {
	CreateThread(threadID uint64) error
}
