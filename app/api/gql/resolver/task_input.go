package resolver

import (
	"github.com/graph-gophers/graphql-go"
)

type TaskInput struct {
	Goal        *string
	DueAt       *graphql.Time
	Context     *string
	OwnerUserID *graphql.ID
}
