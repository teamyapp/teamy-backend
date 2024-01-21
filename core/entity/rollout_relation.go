package entity

type AppRolloutRelationType string

const (
	AppRolloutRelationTypeUser AppRolloutRelationType = "USER"
	AppRolloutRelationTypeTeam AppRolloutRelationType = "TEAM"
)

type AppRolloutRelation struct {
	AppID     uint64
	RolloutID uint64
	Type      AppRolloutRelationType
}

type GroupRolloutRelation struct {
	GroupID    uint64
	RolloutID  uint64
	OrderIndex int
}
