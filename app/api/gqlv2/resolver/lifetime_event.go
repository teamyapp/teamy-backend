package resolver

import (
	"time"

	"github.com/graph-gophers/graphql-go"
)

/////////////////////
// Lifetime Events //
/////////////////////
type LifetimeEventEnum string

const (
	Creation    LifetimeEventEnum = "Creation"
	AssignOwner LifetimeEventEnum = "AssignOwner"
)

type LifetimeEvent struct {
	Deps       *Dependencies
	ID         graphql.ID
	ActorID    graphql.ID
	HappensAt_ time.Time
	// GraphQL fields
	EventType LifetimeEventType
}

func (e LifetimeEvent) Actor() (User, error) {
	return e.Deps.Data.GetUser(e.ActorID)
}

func (e LifetimeEvent) HappensAt() graphql.Time {
	t := graphql.Time{}
	t.Time = e.HappensAt_
	return t
}

type LifetimeEventType struct {
	dep         *Dependencies
	Type        LifetimeEventEnum
	Creation    *EventCreation
	AssignOwner *EventAssignOwner
}

func (e *LifetimeEventType) Dep(dep *Dependencies) {
	e.dep = dep
}

func (e LifetimeEventType) ToCreation() (*EventCreation, bool) {
	if e.Creation != nil {
		e.Creation.dep = e.dep
	}
	return e.Creation, e.Creation != nil
}

func (e LifetimeEventType) ToAssignOwner() (*EventAssignOwner, bool) {
	if e.AssignOwner != nil {
		e.AssignOwner.dep = e.dep
	}
	return e.AssignOwner, e.AssignOwner != nil
}

type EventCreation struct {
	dep    *Dependencies
	TaskID graphql.ID
}

func (e EventCreation) Task() (Task, error) {
	return e.dep.Data.GetTask(e.TaskID)
}

type EventAssignOwner struct {
	dep     *Dependencies
	ownerID graphql.ID
}

func (e EventAssignOwner) Owner() (User, error) {
	return e.dep.Data.GetUser(e.ownerID)
}
