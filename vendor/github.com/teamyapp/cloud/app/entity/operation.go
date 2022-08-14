package entity

import "time"

type Operation struct {
	ResourceTypeName string
	OperationName    string
	CreatedAt        time.Time
	CreatorUserID    uint64
}
