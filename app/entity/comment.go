package entity

import (
	"github.com/graph-gophers/graphql-go"
	oneEntity "github.com/teamyapp/one/entity"
)

type Comment struct {
	ID          oneEntity.ID
	Content     string
	CommenterID graphql.ID
	TaskID      graphql.ID
}
