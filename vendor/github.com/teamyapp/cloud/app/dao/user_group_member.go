package dao

type UserGroupMember interface {
	FindGroupIDsByUserID(userID uint64) ([]uint64, error)
}
