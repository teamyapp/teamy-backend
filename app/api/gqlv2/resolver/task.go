package resolver

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

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
	fmt.Println("task", t.deps)
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
