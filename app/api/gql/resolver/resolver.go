package resolver

import (
	"github.com/teamyapp/teamy-backend/app/service"
)

type Resolver struct {
	Query
	Mutation
}

func NewResolver(executionService service.Execution) Resolver {
	return Resolver{
		Query: NewQuery(executionService),
	}
}
