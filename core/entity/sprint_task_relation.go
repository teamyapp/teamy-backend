package entity

import (
	"time"
)

type SprintTaskRelation struct {
	SprintID  uint64
	TaskID    uint64
	CreatedAt time.Time
}
