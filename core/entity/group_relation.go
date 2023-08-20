package entity

type AppGroupRelationType string

const (
	AppGroupRelationTypeUser AppGroupRelationType = "user"
	AppGroupRelationTypeTeam AppGroupRelationType = "team"
)

type AppGroupRelation struct {
	AppID   uint64
	GroupID uint64
	Type    AppGroupRelationType
}

type UserGroupRelation struct {
	UserID  uint64
	GroupID uint64
}

type TeamGroupRelation struct {
	TeamID  uint64
	GroupID uint64
}
