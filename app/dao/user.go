package dao

type User interface {
	FindUser(id uint64) (User, error)
}
