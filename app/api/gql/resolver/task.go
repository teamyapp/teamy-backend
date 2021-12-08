package resolver

import (
	"fmt"
	"log"
	"strconv"
	"strings"

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
	ID        *graphql.ID
	CreatorID *graphql.ID
	OwnerID   *graphql.ID
	Text      *string
	Status    *entity.TaskStatusEnum
}

func (t Task) Goal() string {
	return t.task.Goal
}

func (t Task) DueAt() *graphql.Time {
	return toGraphQLTime(t.task.DueAt)
}

func (t Task) Context() string {
	return t.task.Context
}

func (t Task) Owner() (*User, error) {
	if t.task.OwnerUserId == nil {
		return nil, nil
	}

	user, err := t.deps.Data.GetUser(*t.task.OwnerUserId)
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

	user, err := t.deps.Data.GetUser(oneEntity.ID(id))
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

func (t Task) AvailableActions() []entity.TaskAction {
	return availableActions[t.task.Status]
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

func (t Task) Comments() []Comment {
	cs := t.deps.Data.FilterComments(func(c entity.Comment) bool {
		return c.TaskID == t.ID()
	})
	return Comments(t.deps, cs)
}

func (t Task) Status() (entity.TaskStatusEnum, error) {
	// TODO: add status to task
	return t.task.Status, nil
}

func newTask(deps *Dependencies, task entity.Task) Task {
	return Task{
		Entity: Entity{entity: task.Entity},
		deps:   deps,
		task:   task,
	}
}

func newTasks(deps *Dependencies, tasks []entity.Task) (ts []Task) {
	for _, t := range tasks {
		ts = append(ts, newTask(deps, t))
	}
	return
}

func taskFilterFunc(t entity.Task, input *TaskFilter) bool {
	if input == nil {
		return true
	}
	// filter by Creator
	matchCreator := true
	if input.CreatorID != nil {
		creatorID := *input.CreatorID
		ownerID := oneEntity.ID(-1)
		if t.OwnerUserId != nil {
			ownerID = *t.OwnerUserId
		}
		matchCreator = t.CreatorID == creatorID || toGraphQLID(ownerID) == creatorID
		log.Println(matchCreator)
	}
	// filter by Owner
	matchOwner := true
	if input.OwnerID != nil && t.OwnerUserId != nil {
		ownerID := *input.OwnerID
		id, err := fromGraphQLID(ownerID)
		if err != nil {
			return false
		}
		matchOwner = (*t.OwnerUserId) == id
	}
	// filter by status
	matchStatus := true
	if input.Status != nil {
		status := *(input.Status)
		matchStatus = t.Status == status
		log.Println("matchStatus", matchStatus)
	}
	// full text search
	// todo: need to implement a better full text search
	// by using the full-text search engine in postgres
	matchText := true
	if input.Text != nil {
		text := *(input.Text)
		matchGoal := strings.Contains(t.Goal, text)
		matchContext := strings.Contains(t.Context, text)
		matchText = matchGoal || matchContext
		log.Println(matchText)
	}
	log.Println("=", matchCreator, matchStatus, matchText)
	return matchCreator && matchStatus && matchText && matchOwner
}
