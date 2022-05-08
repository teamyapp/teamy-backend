package resolver

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type TaskFilter struct {
	OwnerID *graphql.ID
	Goal    *string
	Status  *entityv2.TaskStatus
}

type TeamFilter struct {
	ID *graphql.ID
}

type InvitationFilter struct {
	ID   *graphql.ID
	Code *string
}
