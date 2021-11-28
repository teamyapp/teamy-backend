package resolver

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/graph-gophers/graphql-go"
)

//go:embed graphiql.html
var graphiql []byte

func QraphiQL() []byte {
	return graphiql
}

type Dependencies struct {
	Data *Data
}

type Root struct {
	Deps Dependencies
}

type Mentionable struct {
	dep  Dependencies
	Type string
	ID   int32
}

func (m Mentionable) ToUser() (*User, bool) {
	fmt.Println("m.dep", m.dep)
	u, err := m.dep.Data.GetUser(m.ID)
	if err != nil || m.Type != "User" {
		return nil, false
	}
	return &u, true
}

func (m Mentionable) ToTask() (*Task, bool) {
	tasks := m.dep.Data.GetTasks([]int32{m.ID})
	if len(tasks) == 0 || m.Type != "Task" {
		return nil, false
	}
	return &tasks[0], true
}

type Comment struct{}

func (c Comment) Commenter() User {
	return User{}
}

func (c Comment) Content() string {
	return ""
}
func (t Comment) Mentioned() []Mentionable {
	fmt.Println("comment")
	return []Mentionable{}
}

//////////
// User //
//////////
type User struct {
	deps       Dependencies
	ID         int32
	Name       string
	ProfileUrl string
}

func (u User) Tasks() []Task {
	tasks := u.deps.Data.FilterTasks(func(t Task) bool {
		return t.CreatorID == u.ID
	})
	for i := range tasks {
		tasks[i].deps = u.deps
	}
	return tasks
}

func (u User) TaskNeedAttention() *Task {
	return &Task{}
}

func (u User) UpcomingTasks() []Task {
	return nil
}

func (u User) DeliveredTasks() []Task {
	return nil
}

func (t User) LifetimeEvents() []LifetimeEvent {
	events := t.deps.Data.FilterLifetimeEvents(func(e LifetimeEvent) bool {
		return e.ActorID == t.ID
	})
	for i := range events {
		events[i].deps = t.deps
		events[i].EventType.dep = t.deps
	}
	return events
}

/////////////////////
// Lifetime Events //
/////////////////////
type LifetimeEventEnum string

const (
	Creation    LifetimeEventEnum = "Creation"
	AssignOwner LifetimeEventEnum = "AssignOwner"
)

type LifetimeEvent struct {
	deps       Dependencies
	ID         int32
	ActorID    int32
	HappensAt_ time.Time
	// GraphQL fields
	EventType LifetimeEventType
}

func (e LifetimeEvent) Actor() (User, error) {
	return e.deps.Data.GetUser(e.ActorID)
}

func (e LifetimeEvent) HappensAt() graphql.Time {
	t := graphql.Time{}
	t.Time = e.HappensAt_
	return t
}

type LifetimeEventType struct {
	dep         Dependencies
	Type        LifetimeEventEnum
	Creation    *EventCreation
	AssignOwner *EventAssignOwner
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
	dep    Dependencies
	TaskID int32
}

func (e EventCreation) Task() (Task, error) {
	return e.dep.Data.GetTask(e.TaskID)
}

type EventAssignOwner struct {
	dep     Dependencies
	ownerID int32
}

func (e EventAssignOwner) Owner() (User, error) {
	return e.dep.Data.GetUser(e.ownerID)
}
