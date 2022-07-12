package entity

type ResourceRelation struct {
	ChileResourceID    uint64
	ChildResourceType  string
	ParentResourceID   uint64
	ParentResourceType string
}
