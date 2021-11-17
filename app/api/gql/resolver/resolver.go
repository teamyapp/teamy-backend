package resolver

type Resolver struct {
	Query
	Mutation
}

func NewResolver(query Query, mutation Mutation) Resolver {
	return Resolver{
		Query:    query,
		Mutation: mutation,
	}
}
