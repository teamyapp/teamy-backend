package resolver

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type TaskInput struct {
	Goal        *string
	DueAt       *graphql.Time
	Context     *string
	OwnerUserID *graphql.ID
	Status      *entity.TaskStatusEnum
}
