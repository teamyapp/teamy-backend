package entity

type PermissionQuery struct {
	ResourceID   uint64
	ResourceType string
	Operation    string
	GroupID      uint64
}
