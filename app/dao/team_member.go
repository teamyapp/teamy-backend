package dao

type TeamMember interface {
	FindTeamIDsByUserID(userID uint64) ([]uint64, error)
	FindTeamMemberIDsByTeamID(teamID uint64) ([]uint64, error)
	AddMemberToTeam(userID uint64, teamID uint64) error
	RemoveMemberFromTeam(userID uint64, teamID uint64) error
}
