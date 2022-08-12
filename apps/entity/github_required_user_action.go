package entity

import (
	"time"
)

type GithubRequiredUserAction struct {
	ID                uint64               `json:"id"`
	TeamID            uint64               `json:"teamID"`
	ActionUserID      uint64               `json:"actionUserID"`
	UserActionType    GithubUserActionType `json:"userActionType"`
	IsCompleted       bool                 `json:"isCompleted"`
	RequestedAt       time.Time            `json:"requestedAt"`
	RequestedByUserID uint64               `json:"requestedByUserID"`
}
