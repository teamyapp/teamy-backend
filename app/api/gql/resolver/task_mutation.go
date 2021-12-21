package resolver

import (
	"context"
	"fmt"
	"log"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/teamyapp/one/identity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

func (m Mutation) CreateTask(
	ctx context.Context,
	args struct {
		Task TaskInput
	},
) (Task, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}
	// find user
	user, err := m.deps.Data.GetUser(userID)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	task, err := fromGraphQLTaskInput(args.Task)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	activeTeams := m.deps.Data.FilterTeams(func(t entity.Team) bool {
		return t.ID == user.ActiveTeamID
	})
	if err != nil {
		log.Println(err)
		return Task{}, err
	}
	if len(activeTeams) == 0 {
		return Task{}, fmt.Errorf("user %v does not have an active team", userID)
	}

	task.Status = entity.UPCOMING
	task.OwnerUserId = &userID
	task, err = m.deps.Data.CreateTask(toGraphQLID(userID), activeTeams[0].ID, task)
	if err != nil {
		return Task{}, err
	}
	// add task to team
	_, err = m.deps.Data.UpdateTeam(activeTeams[0].ID, func(t entity.Team) entity.Team {
		t.Tasks = append(t.Tasks, task.ID)
		return t
	})
	if err != nil {
		return Task{}, err
	}
	return newTask(m.deps, task), err
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

	user, err := m.deps.Data.GetUser(userID)
	if err != nil {
		return Task{}, err
	}
	userResolver := User{
		deps: m.deps,
		user: user,
	}
	tasks, err := userResolver.Tasks(struct{ Input *TaskFilter }{})
	if err != nil {
		return Task{}, err
	}

	userCanReachTask := false
	for _, t := range tasks {
		if t.task.ID == task.ID {
			userCanReachTask = true
			break
		}
	}
	allowWrite := false
	{ // the user must be in a team that owns this task
		if userCanReachTask {
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
		} else if task.CreatorID == toGraphQLID(userID) {
			allowWrite = true
		}
	}
	if !allowWrite {
		return Task{}, errors.Errorf("user %v can not modify task %v")
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
