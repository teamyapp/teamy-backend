package dao

import "time"

type TeamMember interface {
	FindTeamIDsByUserID(userID uint64) ([]uint64, error)
	FindTeamMemberIDsByTeamID(teamID uint64) ([]uint64, error)
	CreateTeamMember(teamID uint64, userID uint64) error
	DeleteTeamMember(teamID uint64, userID uint64) error
	UpdateTeamMember(teamID uint64, userID *uint64, needAttentionTaskID *uint64, updatedAt time.Time) error
}
