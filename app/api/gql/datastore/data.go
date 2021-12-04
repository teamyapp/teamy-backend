package datastore

import (
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Data struct {
	Tasks             map[graphql.ID]entity.Task
	Users             map[graphql.ID]entity.User
	Comments          []entity.Comment
	LifetimeEvents    []LifetimeEvent
	CreationRelations []CreationRelation
}

// Temperary SQL like struct for v2 migration purpose.
type CreationRelation struct {
	TaskID graphql.ID
	UserID graphql.ID
}

type LifetimeEvent struct {
	ID         graphql.ID
	ActorID    graphql.ID
	HappensAt_ time.Time
	EventType  LifetimeEventType
}

type LifetimeEventType struct {
	Type        LifetimeEventEnum
	Creation    *EventCreation
	AssignOwner *EventAssignOwner
}

type EventAssignOwner struct {
	ownerID graphql.ID
}

type EventCreation struct {
	TaskID graphql.ID
}

type LifetimeEventEnum string

const (
	Creation    LifetimeEventEnum = "Creation"
	AssignOwner LifetimeEventEnum = "AssignOwner"
)
