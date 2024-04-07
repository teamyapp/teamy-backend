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
	IsScheduled  *bool
	IsPlanned    *bool
}

type TeamFilter struct {
	TeamID *graphql.ID
}

type ProjectFilter struct {
	ProjectID *graphql.ID
}

type PhaseFilter struct {
	PhaseID *graphql.ID
}

type StoryFilter struct {
	StoryID *graphql.ID
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
	TagValues     []string
	IsOnPromotion *bool
}
