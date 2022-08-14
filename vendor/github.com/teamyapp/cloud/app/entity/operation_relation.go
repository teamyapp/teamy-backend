package entity

import "time"

type OperationRelation struct {
	ChildResourceType  string
	ChildOperation     string
	ParentResourceType string
	ParentOperation    string
	CreatedAt          time.Time
	CreatorUserID      uint64
}
