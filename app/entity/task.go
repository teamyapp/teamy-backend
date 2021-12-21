package entity

import (
	"time"

	"github.com/graph-gophers/graphql-go"
	oneEntity "github.com/teamyapp/one/entity"
)

type Task struct {
	oneEntity.Entity
	Goal        string
	DueAt       *time.Time
	Context     string
	CreatorID   graphql.ID
	OwnerUserId *oneEntity.ID
	OwnedByTeam oneEntity.ID
	// WorkScopeIndex   *int
	// Effort           *int
	// DependsOnTaskIDs []oneEntity.ID
	// NumOfUnknowns    *int
	AvailableActions []TaskAction
	Status           TaskStatusEnum
}
