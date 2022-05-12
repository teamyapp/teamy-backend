package datastore

import (
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Data struct {
	Tasks             map[graphql.ID]entity.Task
	Users             map[uint64]entity.User
	Comments          map[uint64]entity.Comment
	LifetimeEvents    []LifetimeEvent
	CreationRelations []CreationRelation
	Teams             []entity.Team
	Invitations       map[uint64]entity.Invitation
	IDs               map[uint64]Type
}

type Type string

const (
	Task       Type = "Task"
	Invitation Type = "Invitation"
	Comment    Type = "Comment"
	Team       Type = "Team"
	Event      Type = "LifetimeEvent"
)

// Temperary SQL like struct for v2 migration purpose.
type CreationRelation struct {
	TaskID graphql.ID
	UserID graphql.ID
}

type LifetimeEvent struct {
	ID         uint64
	ActorID    uint64
	HappensAt_ time.Time
	EventType  LifetimeEventType
}

type LifetimeEventType struct {
	Type        LifetimeEventEnum
	Creation    *EventCreation
	AssignOwner *EventAssignOwner
}

type EventAssignOwner struct {
	OwnerID graphql.ID
}

type EventCreation struct {
	TaskID graphql.ID
}

type LifetimeEventEnum string

const (
	Creation    LifetimeEventEnum = "Creation"
	AssignOwner LifetimeEventEnum = "AssignOwner"
)
