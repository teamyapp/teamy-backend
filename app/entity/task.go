package entity

import (
	"time"

	"github.com/graph-gophers/graphql-go"
)

type Task struct {
	ID          uint64
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	Goal        string
	DueAt       *time.Time
	Context     string
	CreatorID   graphql.ID
	OwnerUserId *uint64
	OwnedByTeam uint64
	// WorkScopeIndex   *int
	// Effort           *int
	// DependsOnTaskIDs []uint64
	// NumOfUnknowns    *int
	AvailableActions []TaskAction
	Status           TaskStatus
}
