package entity

import (
	"time"
)

type GithubRequiredUserAction struct {
	ID                uint64
	TeamID            uint64
	ActionUserID      uint64
	UserActionType    GithubUserActionType
	IsCompleted       bool
	RequestedAt       time.Time
	RequestedByUserID uint64
}
