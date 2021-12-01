package resolver

import (
	_ "embed"

	"github.com/graph-gophers/graphql-go"
)

//go:embed graphiql.html
var graphiql []byte

func QraphiQL() []byte {
	return graphiql
}

type Dependencies struct {
	Data *DataStore
}

func NewDependencies(dataStore *DataStore) *Dependencies {
	return &Dependencies{
		Data: dataStore,
	}
}

type Root struct {
	Deps *Dependencies
}

//////////
// User //
//////////
type User struct {
	deps       *Dependencies
	ID         graphql.ID
	FirstName  string
	LastName   string
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
