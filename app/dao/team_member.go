package dao

type TeamMember interface {
	FindTeamIDsForUser(userId uint64) ([]uint64, error)
}
