package dao

type TeamMember interface {
	FindTeamIDsByUserID(userID uint64) ([]uint64, error)
	FindTeamMemberIDsByTeamID(teamID uint64) ([]uint64, error)
	HasTeamMember(teamID uint64, userID uint64) (bool, error)
	CreateTeamMember(teamID uint64, userID uint64) error
	DeleteTeamMember(teamID uint64, userID uint64) error
}
