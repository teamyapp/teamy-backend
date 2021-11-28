package resolver

import (
	_ "embed"
	"fmt"

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

func NewDependencies() *Dependencies {
	return &Dependencies{
		Data: Read("./data.json"),
	}
}

type Root struct {
	Deps *Dependencies
}

type Mentionable struct {
	dep  *Dependencies
	Type string
	ID   graphql.ID
}

func (m Mentionable) ToUser() (*User, bool) {
	fmt.Println("Mentionable: m.dep", m.dep)
  if m.Type != "User" {
		return nil, false
	}
	u, err := m.dep.Data.GetUser(m.ID)
	if err != nil {
		return nil, false
	}
	return &u, true
}

func (m Mentionable) ToTask() (*Task, bool) {
  if m.Type != "Task" {
    return nil, false
  }
	tasks := m.dep.Data.GetTasks([]graphql.ID{m.ID})
	if len(tasks) == 0 {
		return nil, false
	}
	return &tasks[0], true
}

//////////
// User //
//////////
type User struct {
	deps       *Dependencies
	ID         graphql.ID
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
		events[i].Deps = t.deps
		events[i].EventType.dep = t.deps
	}
	return events
}
