package resolver

type Mutation struct {
	deps  *Dependencies
	query *Query
}

func NewMutation(deps *Dependencies, query *Query) Mutation {
	return Mutation{
		deps:  deps,
		query: query,
	}
}
