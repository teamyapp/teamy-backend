package gql

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskFilter struct {
	TaskID       *graphql.ID
	OwnerID      *graphql.ID
	GoalContains *string
	Status       *entity.TaskStatus
	IsPlanned    *bool
}

type TeamFilter struct {
	TeamID *graphql.ID
}

type InvitationFilter struct {
	InvitationID *graphql.ID
	Code         *string
}

type SprintFilter struct {
	SprintID        *graphql.ID
	StartAtAndAfter *graphql.Time
	SortByStartAt   *bool
	CountLimit      *int32
}

type AppFilter struct {
	Query         *string
	Tag           *string
	IsOnPromotion *bool
}
