package resolver

import (
	"fmt"
	"log"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/api/gql/datastore"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Task struct {
	deps *Dependencies
	Entity
	task entity.Task
}

type TaskFilter struct {
	ID     *graphql.ID
	Text   *string
	Status *TaskStatus
}

func (t Task) Goal() string {
	return t.task.Goal
}

func (t Task) DueAt() *graphql.Time {
	return toGraphQLTime(t.task.DueAt)
}

func (t Task) Context() string {
	c := t.task.Context
	if c != nil {
		return *c
	}
	return ""
}

func (t Task) Owner() (*User, error) {
	if t.task.OwnerUserId == nil {
		return nil, nil
	}

	user, err := t.deps.userRepo.FindUser(*t.task.OwnerUserId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	gqlUser := newUser(t.deps, user)
	return &gqlUser, nil
}

func (t Task) Creator() (User, error) {
	rs := t.deps.Data.FilterCreationRelation(func(cr datastore.CreationRelation) bool {
		return t.ID() == cr.TaskID
	})
	if len(rs) == 0 {
		return User{}, fmt.Errorf("this task %v has no creator recorded", t.ID())
	}
	id, err := strconv.Atoi(string(rs[0].UserID))
	if err != nil {
		return User{}, err
	}

	user, err := t.deps.userRepo.FindUser(oneEntity.ID(id))
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	gqlUser := newUser(t.deps, user)
	return gqlUser, nil
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

func (t Task) LifetimeEvents() []LifetimeEvent {
	events := t.deps.Data.FilterLifetimeEvents(func(e datastore.LifetimeEvent) bool {
		return e.EventType.Creation.TaskID == t.ID()
	})
	return LifetimeEvents(events)
}

func (t Task) Mentions() ([]Mention, error) {
	return ParseMentions(t.Context()), nil
}

type TaskStatus string

const (
	UPCOMING    TaskStatus = "UPCOMING"
	IN_PROGRESS TaskStatus = "IN_PROGRESS"
	DELIVERED   TaskStatus = "DELIVERED"
)

func (t Task) Status() (TaskStatus, error) {
	// TODO: add status to task
	return UPCOMING, nil
}

func newTask(deps *Dependencies, task entity.Task) Task {
	return Task{
		Entity: Entity{entity: task.Entity},
		deps:   deps,
		task:   task,
	}
}
