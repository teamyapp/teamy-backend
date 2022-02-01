package resolver

import (
	"context"
	"log"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/teamyapp/one/identity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

func (m Mutation) TaskUpdate(
	ctx context.Context,
	args struct{ TaskID graphql.ID },
) (TaskUpdate, error) {
	id, err := fromGraphQLID(args.TaskID)
	if err != nil {
		return TaskUpdate{}, err
	}

	tasks := m.deps.Data.FilterTasks(func(t entity.Task) bool {
		return t.ID == id
	})
	if len(tasks) < 1 {
		return TaskUpdate{}, errors.Errorf("task not found: id=%s", args.TaskID)
	}

	return TaskUpdate{
		deps: m.deps,
		task: tasks[0],
	}, nil
}

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
	if args.Task.OwnerUserID != nil {
		id, err := fromGraphQLID(*args.Task.OwnerUserID)
		if err != nil {
			return TaskUpdate{}, err
		}
		task.OwnerUserId = &id
	}

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

func (tu TaskUpdate) OwnedByTeam(ctx context.Context, args struct{ ID graphql.ID }) (TaskUpdate, error) {
	teamID, err := fromGraphQLID(args.ID)
	if err != nil {
		return tu, err
	}
	userID, err := identity.FromContext(ctx)
	if err != nil {
		return TaskUpdate{}, err
	}
	// check if the user is a member of the team
	teams := tu.deps.Data.FilterTeams(func(t entity.Team) bool {
		return t.ID == teamID
	})
	if len(teams) == 0 {
		return tu, errors.Errorf("team %v does not exist", teamID)
	}
	if !teams[0].MemberIDs.Has(userID) {
		return tu, errors.Errorf("user %v is not a member of team %v", userID, teamID)
	}

	// add the task to the team's task list if not there yet
	_, err = tu.deps.Data.UpdateTeam(teamID, func(t entity.Team) entity.Team {
		t.Tasks = t.Tasks.Add(tu.task.ID)
		return t
	})
	if err != nil {
		return tu, err
	}

	// make this team own this task (all team members has UPDATE privilige to the task)
	tu.task.OwnedByTeam = teamID
	newTask, err := tu.deps.Data.UpdateTask(tu.task)
	if err != nil {
		return tu, err
	}
	tu.task = newTask
	return tu, nil
}

func (tu TaskUpdate) RemoveOwner(ctx context.Context) (TaskUpdate, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		return TaskUpdate{}, err
	}

	{ // access control
		user, err := tu.deps.Data.GetUser(userID)
		if err != nil {
			return TaskUpdate{}, err
		}
		err = allowWrite(ctx, tu.task, User{
			deps: tu.deps,
			user: user,
		})
		if err != nil {
			return TaskUpdate{}, err
		}
	}

	tu.task.OwnerUserId = nil
	_, err = tu.deps.Data.UpdateTask(tu.task)
	if err != nil {
		return TaskUpdate{}, err
	}
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
	{ // access control
		user, err := m.deps.Data.GetUser(userID)
		if err != nil {
			return Task{}, err
		}
		err = allowWrite(ctx, task, User{
			deps: m.deps,
			user: user,
		})
		if err != nil {
			return Task{}, err
		}
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
		err = allowWrite(ctx, task, userResolver)
		if err != nil {
			return Task{}, err
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

func allowWrite(ctx context.Context, task entity.Task, userResolver User) error {
	allowWrite := false
	// the user must be in a team that owns this task
	teams, err := userResolver.Teams(ctx, struct{ IDs *[]graphql.ID }{})
	if err != nil {
		return err
	}
	for _, team := range teams {
		if task.OwnedByTeam == team.Team.ID {
			allowWrite = true
			break
		}
	}
	if task.CreatorID == toGraphQLID(userResolver.user.ID) {
		allowWrite = true
	}

	if !allowWrite {
		return errors.Errorf("user %v can not modify task %v", userResolver.user.ID, task.ID)
	}
	return nil
}
