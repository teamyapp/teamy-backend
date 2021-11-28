package resolver

import "github.com/graph-gophers/graphql-go"

type User struct {
	Entity
	ID graphql.ID
}

func (u User) Name() string {
	panic("not implemented")
}

func (u User) ProfileURL() string {
	panic("not implemented")
}
