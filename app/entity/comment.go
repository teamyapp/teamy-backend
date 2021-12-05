package entity

import "github.com/graph-gophers/graphql-go"

type Comment struct {
	ID          graphql.ID
	Content     string
	CommenterID graphql.ID
	TaskID      graphql.ID
}
