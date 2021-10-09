package resolver

import (
	"github.com/graph-gophers/graphql-go"
)

type Entity struct {
}

func (e Entity) ID() graphql.ID {
	panic("not implemented")
}

func (e Entity) CreatedAt() graphql.Time {
	panic("not implemented")
}

func (e Entity) UpdatedAt() *graphql.Time {
	panic("not implemented")
}
