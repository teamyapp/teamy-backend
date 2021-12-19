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

type Mutation struct {
	deps  *Dependencies
	query *Query
}

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

// func (m Mutation) StartTask(ctx context.Context, args struct {
// 	TaskID graphql.ID
// }) (bool, error) {
// 	userID, err := identity.FromContext(ctx)
// 	if err != nil {
// 		log.Println(err)
// 		return false, err
// 	}

// 	taskID, err := fromGraphQLID(args.TaskID)
// 	if err != nil {
// 		log.Println(err)
// 		return false, err
// 	}

// 	// TODO: a user starts others' task will assign that task to the himself
// 	// TODO: show a modal to confirm task should be reassigned.
// 	activeTeam, err := m.deps.teamRepo.FindActiveTeam(userID)
// 	if err != nil {
// 		log.Println(err)
// 		return false, err
// 	}

// 	prevNeedAttentionTaskID, err := m.deps.taskRepo.SetNeedAttentionTask(&taskID, userID, activeTeam.ID)
// 	if err != nil {
// 		log.Println(err)
// 		return false, err
// 	}

// 	err = m.deps.taskRepo.SetTeamTaskStatus(taskID, activeTeam.ID, entity.TaskStatusInProgress)
// 	if err != nil {
// 		log.Println(err)
// 		return false, err
// 	}

// 	if prevNeedAttentionTaskID != nil {
// 		err = m.deps.taskRepo.SetTeamTaskStatus(*prevNeedAttentionTaskID, activeTeam.ID, entity.TaskStatusUpcoming)
// 		if err != nil {
// 			log.Println(err)
// 			return false, err
// 		}
// 	}

// 	return true, nil
// }

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

	// taskID, err := fromGraphQLID(args.TaskID)
	// if err != nil {
	// 	log.Println(err)
	// 	return false, err
	// }

	// Check if taskID exists - no need, we can only delete task that we can see
	// UI:
	//		delete task from all active team view
	// Delete record from team_task table
	// TODO: move task to trash instead of completely deleting it. delete task after 7 days if in action
	// TODO: clean up the task from task dependency graph for the active team

	// question: not sure if this is useful anymore
	// activeTeam, err := m.deps.teamRepo.FindActiveTeam(userID)
	// if err != nil {
	// 	log.Println(err)
	// 	return false, err
	// }

	return newTask(m.deps, deletedTask), nil
}

// func (m Mutation) CompleteTask(ctx context.Context, args struct {
// 	TaskID graphql.ID
// }) (bool, error) {
// 	userID, err := identity.FromContext(ctx)
// 	if err != nil {
// 		log.Println(err)
// 		return false, err
// 	}

// 	taskID, err := fromGraphQLID(args.TaskID)
// 	if err != nil {
// 		log.Println(err)
// 		return false, err
// 	}

// 	activeTeam, err := m.deps.teamRepo.FindActiveTeam(userID)
// 	if err != nil {
// 		log.Println(err)
// 		return false, err
// 	}

// 	needAttentionTask, err := m.deps.taskRepo.FindTaskNeedAttentionForUser(userID, activeTeam.ID)
// 	if err != nil {
// 		log.Println(err)
// 		return false, err
// 	}

// 	if needAttentionTask.ID != taskID {
// 		return true, err
// 	}

// 	_, err = m.deps.taskRepo.SetNeedAttentionTask(nil, userID, activeTeam.ID)
// 	if err != nil {
// 		log.Println(err)
// 		return false, err
// 	}

// 	err = m.deps.taskRepo.SetTeamTaskStatus(taskID, activeTeam.ID, entity.TaskStatusDelivered)
// 	if err != nil {
// 		log.Println(err)
// 		return false, err
// 	}

// 	return true, nil
// }

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
	if task.CreatorID != toGraphQLID(userID) {
		return Task{}, errors.New("you are not the creator of this task")
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

func (m Mutation) Comment(
	ctx context.Context,
	args struct {
		TaskID  graphql.ID
		Content string
	},
) (Comment, error) {
	// partial working
	userID, err := identity.FromContext(ctx)
	if err != nil {
		return Comment{}, err
	}
	c, err := m.deps.Data.CreateComment(entity.Comment{
		Content:     args.Content,
		CommenterID: toGraphQLID(userID),
		TaskID:      args.TaskID,
	})
	if err != nil {
		return Comment{}, err
	}
	return Comment{
		deps:    m.deps,
		Comment: c,
	}, nil
}

func NewMutation(deps *Dependencies, query *Query) Mutation {
	return Mutation{
		deps:  deps,
		query: query,
	}
}
