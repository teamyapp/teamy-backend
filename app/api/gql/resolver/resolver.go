package resolver

import "github.com/teamyapp/teamy-backend/app/api/gqlv2/resolver"

type Resolver struct {
	Query
	Mutation
}

func NewResolver(query Query, mutation Mutation, dep *resolver.Dependencies) Resolver {
	query.dep = dep
	mutation.dep = dep
	return Resolver{
		Query:    query,
		Mutation: mutation,
	}
}
