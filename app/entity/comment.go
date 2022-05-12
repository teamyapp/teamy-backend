package entity

import (
	"time"

	"github.com/graph-gophers/graphql-go"
)

type Comment struct {
	ID          uint64
	Content     string
	CommenterID graphql.ID
	TaskID      graphql.ID
	CreatedAt   time.Time
}
