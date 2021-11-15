package resolver

type Resolver struct {
	Query
	Mutation
}

func NewResolver(query Query) Resolver {
	return Resolver{
		Query: query,
	}
}
