package resolver

import (
	"context"
	"log"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/teamyapp/one/identity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

func (m Mutation) CreateTask(
	ctx context.Context,
	args struct{ Task TaskInput },
) (TaskUpdate, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		return TaskUpdate{}, err
	}

	task, err := fromGraphQLTaskInput(args.Task)
	if err != nil {
		return TaskUpdate{}, err
	}

	task.Status = entity.UPCOMING
	task.OwnerUserId = &userID
	task, err = m.deps.Data.CreateTask(toGraphQLID(userID), task)
	if err != nil {
		return TaskUpdate{}, err
	}
	return TaskUpdate{m.deps, task}, err
}

type TaskUpdate struct {
	deps *Dependencies
	task entity.Task
}

func (tu TaskUpdate) OwnedByTeam(args struct{ ID graphql.ID }) (TaskUpdate, error) {
	id, err := fromGraphQLID(args.ID)
	if err != nil {
		return tu, err
	}
	tu.task.OwnedByTeam = id
	newTask, err := tu.deps.Data.UpdateTask(tu.task)
	if err != nil {
		return tu, err
	}
	tu.task = newTask
	return tu, nil
}

func (tu TaskUpdate) Task() Task {
	return newTask(tu.deps, tu.task)
}

func (m Mutation) DeleteTask(ctx context.Context, args struct {
	TaskID graphql.ID
}) (Task, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	// can only delete a task if
	// the user is the creator
	// todo: need to design a fully featured authorization interface for all modifiable entities
	task, err := m.deps.Data.GetTask(args.TaskID)
	if err != nil {
		return Task{}, err
	}
	if task.CreatorID != toGraphQLID(userID) {
		return Task{}, errors.New("you are not the creator of this task, can not delete")
	}

	deletedTask, err := m.deps.Data.DeleteTask(args.TaskID)
	if err != nil {
		return Task{}, err
	}

	return newTask(m.deps, deletedTask), nil
}

func (m Mutation) UpdateTask(
	ctx context.Context,
	args struct {
		TaskID graphql.ID
		Task   TaskInput
	},
) (Task, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}
	task, err := m.deps.Data.GetTask(args.TaskID)
	if err != nil {
		return Task{}, err
	}
	{ // access control
		user, err := m.deps.Data.GetUser(userID)
		if err != nil {
			return Task{}, err
		}
		userResolver := User{
			deps: m.deps,
			user: user,
		}
		allowWrite := false
		{ // the user must be in a team that owns this task
			teams, err := userResolver.Teams(ctx, struct{ IDs *[]graphql.ID }{})
			if err != nil {
				return Task{}, err
			}
			for _, team := range teams {
				if task.OwnedByTeam == team.Team.ID {
					allowWrite = true
					break
				}
			}
			if task.CreatorID == toGraphQLID(userID) {
				allowWrite = true
			}
		}
		if !allowWrite {
			return Task{}, errors.Errorf("user %v can not modify task %v", userID, task.ID)
		}
	}
	if args.Task.Context != nil {
		task.Context = *(args.Task.Context)
	}
	if args.Task.DueAt != nil {
		task.DueAt = &args.Task.DueAt.Time
	}
	if args.Task.Goal != nil {
		task.Goal = *args.Task.Goal
	}
	if args.Task.OwnerUserID != nil {
		id, err := fromGraphQLIDPtr(args.Task.OwnerUserID)
		if err != nil {
			return Task{}, nil
		}
		task.OwnerUserId = id
	}
	if args.Task.Status != nil {
		log.Println(*args.Task.Status)
		task.Status = *args.Task.Status
	}
	task, err = m.deps.Data.UpdateTask(task)
	if err != nil {
		return Task{}, nil
	}
	return newTask(m.deps, task), nil
}
