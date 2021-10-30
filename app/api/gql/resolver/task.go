package resolver

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Task struct {
	Entity
	task entity.Task
}

func (t Task) Goal() string {
	return t.task.Goal
}

func (t Task) DueAt() *graphql.Time {
	return toGraphQLTime(t.task.DueAt)
}

func (t Task) Context() *string {
	return t.task.Context
}

func (t Task) Owner() *User {
	panic("not implemented")
}

func (t Task) WorkScope() Option {
	panic("not implemented")
}

func (t Task) Effort() *int32 {
	return toGraphQLInt(t.task.Effort)
}

func (t Task) DependsOn() []Task {
	panic("not implemented")
}

func (t Task) NumOfUnknowns() *int32 {
	return toGraphQLInt(t.task.NumOfUnknowns)
}

func (t Task) AvailableActions() []TaskAction {
	return toGraphQLTaskActions(t.task.AvailableActions)
}

func (t Task) AvailableWorkScopes() []Option {
	panic("not implemented")
}

func newTask(task entity.Task) Task {
	return Task{
		Entity: Entity{entity: task.Entity},
		task:   task,
	}
}
