package resolver

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type TaskFilter struct {
	ID      *graphql.ID
	OwnerID *graphql.ID
	Goal    *string
	Status  *entity.TaskStatus
}

type TeamFilter struct {
	ID *graphql.ID
}

type InvitationFilter struct {
	ID   *graphql.ID
	Code *string
}
