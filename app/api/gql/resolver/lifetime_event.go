package resolver

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/api/gql/datastore"
)

type LifetimeEvent struct {
	datastore.LifetimeEvent
}

func (e LifetimeEvent) HappensAt() graphql.Time {
	return graphql.Time{Time: e.HappensAt_}
}

func LifetimeEvents(es []datastore.LifetimeEvent) (events []LifetimeEvent) {
	for _, e := range es {
		events = append(events, LifetimeEvent{
			e,
		})
	}
	return
}
