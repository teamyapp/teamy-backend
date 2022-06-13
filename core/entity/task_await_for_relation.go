package entity

import (
	"time"
)

type TaskAwaitForRelation struct {
	AWaitingTaskID uint64
	AWaitForTaskID uint64
	CreatedAt      time.Time
}
