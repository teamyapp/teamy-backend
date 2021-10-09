package resolver

import (
	"github.com/graph-gophers/graphql-go"
)

type Task struct {
	Entity
}

func (t Task) Goal() string {
	panic("not implemented")
}

func (t Task) DueAt() *graphql.Time {
	panic("not implemented")
}

func (t Task) Context() *string {
	panic("not implemented")
}

func (t Task) Owner() *User {
	panic("not implemented")
}

func (t Task) WorkScope() Option {
	panic("not implemented")
}

func (t Task) Effort() *int32 {
	panic("not implemented")
}

func (t Task) DependsOn() []Task {
	panic("not implemented")
}

func (t Task) NumOfUnknowns() *int32 {
	panic("not implemented")
}

func (t Task) AvailableActions() []TaskAction {
	panic("not implemented")
}

func (t Task) AvailableWorkScopes() []Option {
	panic("not implemented")
}
