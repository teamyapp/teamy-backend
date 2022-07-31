package entity

import (
	"time"
)

type TaskAwaitForRelation struct {
	AwaitingTaskID uint64
	AwaitForTaskID uint64
	CreatedAt      time.Time
}
