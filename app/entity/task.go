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
	OwnerUserId      *int
	WorkScopeIndex   int
	Effort           *int
	DependsOnTaskIDs []int
	NumOfUnknowns    *int
}
