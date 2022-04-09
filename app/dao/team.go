package dao

type Team interface {
	FindTeam(id uint64) (Team, error)
}
