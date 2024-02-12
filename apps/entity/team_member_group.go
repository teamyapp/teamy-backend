package entity

type TeamMemberGroup struct {
	ID            uint64   `json:"id"`
	TeamID        uint64   `json:"teamId"`
	Name          string   `json:"name"`
	MemberUserIDs []uint64 `json:"memberUserIds"`
}
