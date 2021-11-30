package resolver

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/api/gqlv2/resolver"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Task struct {
	deps          *Dependencies
	prototypeDeps *resolver.Dependencies
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

func (t Task) Owner() (*User, error) {
	userID := t.task.OwnerUserId
	if userID == nil {
		return nil, nil
	}

	return nil, errors.New("not implemented")
}

func (t Task) Creator() (User, error) {
	rs := t.prototypeDeps.Data.FilterCreationRelation(func(cr resolver.CreationRelation) bool {
		return t.ID() == cr.TaskID
	})
	if len(rs) == 0 {
		return User{}, fmt.Errorf("this task %v has no creator recorded", t.ID())
	}
	id, err := strconv.Atoi(string(rs[0].UserID))
	return User{
		Entity: Entity{
			entity: oneEntity.Entity{
				ID: oneEntity.ID(id),
			},
		},
	}, err
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

func (t Task) LifetimeEvents() []resolver.LifetimeEvent {
	events := t.prototypeDeps.Data.FilterLifetimeEvents(func(e resolver.LifetimeEvent) bool {
		return e.EventType.Creation.TaskID == t.ID()
	})
	for i := range events {
		events[i].Deps = t.prototypeDeps
		events[i].EventType.Dep(t.prototypeDeps)
	}
	return events
}

func (t Task) Mentions() ([]resolver.Mention, error) {
	if t.Context() == nil {
		return nil, nil
	}
	return resolver.ParseMentions(*t.Context()), nil
}

func newTask(deps *Dependencies, prototypeDeps *resolver.Dependencies, task entity.Task) Task {
	return Task{
		Entity:        Entity{entity: task.Entity},
		deps:          deps,
		prototypeDeps: prototypeDeps,
		task:          task,
	}
}
