package resolver

import (
	"fmt"

	"github.com/graph-gophers/graphql-go"
	oneEntity "github.com/teamyapp/one/entity"
)

type Entity struct {
	entity oneEntity.Entity
}

func (e Entity) ID() graphql.ID {
	return graphql.ID(fmt.Sprintf("%d", int(e.entity.ID)))
}

func (e Entity) CreatedAt() graphql.Time {
	return graphql.Time{Time: e.entity.CreatedAt}
}

func (e Entity) UpdatedAt() *graphql.Time {
	return toGraphQLTime(e.entity.UpdatedAt)
}
