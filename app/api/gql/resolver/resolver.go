package resolver

type Resolver struct {
	Query
	Mutation
}

func NewResolver(deps *Dependencies) Resolver {
	return Resolver{
		Query:    NewQuery(deps),
		Mutation: NewMutation(deps),
	}
}
