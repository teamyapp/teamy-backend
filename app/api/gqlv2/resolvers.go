package gqlv2

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
)

//go:embed graphiql.html
var graphiql []byte

func QraphiQL() []byte {
	return graphiql
}

//go:embed schema.gql
var rawSchema string

func RawSchema() string {
	return rawSchema
}

type Dependencies struct {
	Data *Data
}

type Root struct {
	Deps Dependencies
}

func (r Root) Tasks(args struct{ ID int32 }) ([]Task, error) {
	tasks := r.Deps.Data.GetTasks([]int32{args.ID})
	for i := range tasks {
		tasks[i].deps = r.Deps
	}
	return tasks, nil
}

func (r Root) Me() (User, error) {
	u, err := r.Deps.Data.GetUser(1)
	u.deps = r.Deps
	return u, err
}

type TaskInput struct {
	Goal    string
	Context *string
}

func (r Root) CreateTask(args struct{ Input TaskInput }) (Task, error) {
	return r.Deps.Data.CreateTask(args.Input, 1)
}

type Task struct {
	// this is always injected by a parent level resolver which returns a Task, at runtime
	deps Dependencies
	// these 4 can be fetched by a parent level resolver which returns a Task
	// in terms of SQL, they could be part of the Task table
	// other method resolvers could be "inner joins"
	ID      int32
	Goal    string
	Context string
	DueAt   *time.Time

	// foreign keys
	CreatorID int32
}

// Mentioned could be a function of Goal and Context
func (t Task) Mentioned() []Mentionable {
	parseMentioned := func(input string) (m []Mentionable) {
		chunks := strings.Split(input, " ")
		for _, chunk := range chunks {
			if len(chunk) == 0 {
				continue
			}
			if chunk[0] == '@' {
				id, err := strconv.ParseInt(chunk[1:], 10, 32)
				if err != nil {
					continue
				}
				m = append(m, Mentionable{
					dep:  t.deps,
					Type: "User",
					ID:   int32(id),
				})
			} else if chunk[0] == '#' {
				id, err := strconv.ParseInt(chunk[1:], 10, 32)
				if err != nil {
					continue
				}
				m = append(m, Mentionable{
					dep:  t.deps,
					Type: "Task",
					ID:   int32(id),
				})
			}
		}
		return
	}
	return parseMentioned(t.Context)
}
func (t Task) Comments() []Comment { return []Comment{} }
func (t Task) DependsOn() []Task   { return []Task{} }
func (t Task) Creator() (User, error) {
	user, err := t.deps.Data.GetUser(t.CreatorID)
	user.deps = t.deps
	return user, err
}
func (t Task) Assignees() []User { return []User{} }
func (t Task) LifetimeEvents() []LifetimeEvent {
	events := t.deps.Data.FilterLifetimeEvents(func(e LifetimeEvent) bool {
		fmt.Println(e.EventType.Creation.TaskID, t.ID)
		return e.EventType.Creation.TaskID == t.ID
	})
	for i := range events {
		events[i].deps = t.deps
		events[i].EventType.dep = t.deps
	}
	return events
}

type Mentionable struct {
	dep  Dependencies
	Type string
	ID   int32
}

func (m Mentionable) ToUser() (*User, bool) {
	u, err := m.dep.Data.GetUser(m.ID)
	if err != nil {
		return nil, false
	}
	return &u, true
}

func (m Mentionable) ToTask() (*Task, bool) {
	tasks := m.dep.Data.GetTasks([]int32{m.ID})
	if len(tasks) == 0 {
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
