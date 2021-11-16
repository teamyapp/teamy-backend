package entity

import (
	"time"

	oneEntity "github.com/teamyapp/one/entity"
)

type Task struct {
	oneEntity.Entity
	Goal             string
	DueAt            *time.Time
	Context          *string
	OwnerUserId      *oneEntity.ID
	WorkScopeIndex   *int
	Effort           *int
	DependsOnTaskIDs []oneEntity.ID
	NumOfUnknowns    *int
	AvailableActions []TaskAction
}
