package entity

import (
	"time"
)

type Task struct {
	Entity
	Goal             string
	DueAt            *time.Time
	Context          string
	OwnerUserId      int
	WorkScopeIndex   int
	Effort           int
	DependsOnTaskIDs []int
	NumOfUnknowns    int
}
