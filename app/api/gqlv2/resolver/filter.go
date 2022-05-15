package resolver

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type TaskFilter struct {
	TaskID  *graphql.ID
	OwnerID *graphql.ID
	Goal    *string
	Status  *entityv2.TaskStatus
}

type TeamFilter struct {
	TeamID *graphql.ID
}

type InvitationFilter struct {
	InvitationID *graphql.ID
	Code         *string
}
