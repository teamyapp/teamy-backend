package dao

type TeamMember interface {
	FindTeamIDsByUserID(userID uint64) ([]uint64, error)
}
