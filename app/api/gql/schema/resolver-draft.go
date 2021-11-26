package schema

import (
	"time"
)

type Task struct {
	// this is always injected by a parent level resolver which returns a Task, at runtime
	deps           DependencyObject
	// these 4 can be fetched by a parent level resolver which returns a Task
	// in terms of SQL, they could be part of the Task table
	// other method resolvers could be "inner joins" 
	ID             graphql.ID
	Goal           string
	Context        string
	DueAt          time.Time
}

// Mentioned could be a function of Goal and Context
func (t Task) Mentioned()      []Mentionable {}
func (t Task) Comments()       []Comment {}
func (t Task) DependsOn()      []Task {}
func (t Task) Creator()        User   {}
func (t Task) Assignees()      []User {}
func (t Task) LifetimeEvents() []TaskLifetimeEvent {}
