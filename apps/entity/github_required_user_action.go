package entity

import (
	"time"
)

type GithubRequiredUserAction struct {
	TeamID            uint64               `json:"teamID"`
	ActionUserID      uint64               `json:"actionUserID"`
	UserActionType    GithubUserActionType `json:"userActionType"`
	IsCompleted       bool                 `json:"isCompleted"`
	RequestedAt       time.Time            `json:"requestedAt"`
	RequestedByUserID uint64               `json:"requestedByUserID"`
}
