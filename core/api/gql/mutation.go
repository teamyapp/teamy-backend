package gql

type Mutation struct {
	deps *Dependencies
}

func NewMutation(deps *Dependencies) Mutation {
	return Mutation{
		deps: deps,
	}
}
