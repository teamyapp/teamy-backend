package resolver

import "github.com/teamyapp/teamy-backend/app/api/gqlv2/resolver"

type Resolver struct {
	Query
	Mutation
}

func NewResolver(deps *Dependencies, prototypeDeps *resolver.Dependencies) Resolver {
	query := NewQuery(deps, prototypeDeps)
	return Resolver{
		Query:    NewQuery(deps, prototypeDeps),
		Mutation: NewMutation(deps, prototypeDeps, &query),
	}
}
