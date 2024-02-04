package entity

type AppGroupRelation struct {
	AppID   uint64
	GroupID uint64
}

type GroupMemberRelation struct {
	GroupID  uint64
	MemberID uint64
}
