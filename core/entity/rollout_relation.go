package entity

type AppRolloutRelationType string

const (
	AppRolloutRelationTypeUser AppRolloutRelationType = "user"
	AppRolloutRelationTypeTeam AppRolloutRelationType = "team"
)

type AppRolloutRelation struct {
	AppID     uint64
	RolloutID uint64
	Type      AppRolloutRelationType
}

type RolloutVersionRelation struct {
	RolloutID     uint64
	VersionNumber int
}

type GroupRolloutRelation struct {
	GroupID    uint64
	RolloutID  uint64
	OrderIndex int
}
