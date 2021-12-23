package datastore

import (
	"time"

	"github.com/graph-gophers/graphql-go"
	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Data struct {
	Tasks             map[graphql.ID]entity.Task
	Users             map[oneEntity.ID]entity.User
	Comments          map[oneEntity.ID]entity.Comment
	LifetimeEvents    []LifetimeEvent
	CreationRelations []CreationRelation
	Teams             []entity.Team
	IDs               map[oneEntity.ID]Type
}

type Type string

const (
	Task    Type = "Task"
	Comment Type = "Comment"
	Team    Type = "Team"
	Event   Type = "LifetimeEvent"
)

// Temperary SQL like struct for v2 migration purpose.
type CreationRelation struct {
	TaskID graphql.ID
	UserID graphql.ID
}

type LifetimeEvent struct {
	ID         oneEntity.ID
	ActorID    oneEntity.ID
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
