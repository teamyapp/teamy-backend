package resolver

import (
	"context"
	"fmt"
	"log"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/one/identity"
	"github.com/teamyapp/teamy-backend/app/api/gqlv2/resolver"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Mutation struct {
	deps          *Dependencies
	prototypeDeps *resolver.Dependencies
	query         *Query
}

func (m Mutation) CreateTask(ctx context.Context, args struct {
	Task TaskInput
}) (Task, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	task, err := fromGraphQLTaskInput(args.Task)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	activeTeam, err := m.deps.teamRepo.FindActiveTeam(userID)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}
	if activeTeam == nil {
		return Task{}, fmt.Errorf("user %v does not have an active team", userID)
	}

	taskID, err := m.deps.taskRepo.CreateTask(task)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	err = m.deps.taskRepo.AssignTaskToTeam(taskID, activeTeam.ID, entity.TaskStatusUpcoming)

	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	err = m.prototypeDeps.Data.CreateLifetimeEvent(graphql.ID(fmt.Sprint(userID)), resolver.LifetimeEventType{
		Type: resolver.Creation,
		Creation: &resolver.EventCreation{
			TaskID: graphql.ID(fmt.Sprint(taskID)),
		},
	})
	if err != nil {
		log.Println(err)
	}

	gqlTask, err := m.query.Task(
		struct {
			ID graphql.ID
		}{
			ID: toGraphQLID(taskID),
		})
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	m.prototypeDeps.Data.CreationRelations = append(m.prototypeDeps.Data.CreationRelations, resolver.CreationRelation{
		TaskID: gqlTask.ID(),
		UserID: graphql.ID(fmt.Sprint(userID)),
	})

	return gqlTask, nil
}

func (m Mutation) StartTask(ctx context.Context, args struct {
	TaskID graphql.ID
}) (bool, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return false, err
	}

	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	// TODO: a user starts others' task will assign that task to the himself
	// TODO: show a modal to confirm task should be reassigned.
	activeTeam, err := m.deps.teamRepo.FindActiveTeam(userID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	prevNeedAttentionTaskID, err := m.deps.taskRepo.SetNeedAttentionTask(&taskID, userID, activeTeam.ID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	err = m.deps.taskRepo.SetTeamTaskStatus(taskID, activeTeam.ID, entity.TaskStatusInProgress)
	if err != nil {
		log.Println(err)
		return false, err
	}

	if prevNeedAttentionTaskID != nil {
		err = m.deps.taskRepo.SetTeamTaskStatus(*prevNeedAttentionTaskID, activeTeam.ID, entity.TaskStatusUpcoming)
		if err != nil {
			log.Println(err)
			return false, err
		}
	}

	return true, nil
}

func (m Mutation) DeleteTask(ctx context.Context, args struct {
	TaskID graphql.ID
}) (bool, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return false, err
	}

	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	// Check if taskID exists - no need, we can only delete task that we can see
	// UI:
	//		delete task from all active team view
	// Delete record from team_task table
	// TODO: move task to trash instead of completely deleting it. delete task after 7 days if in action
	// TODO: clean up the task from task dependency graph for the active team

	activeTeam, err := m.deps.teamRepo.FindActiveTeam(userID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	err = m.deps.taskRepo.DeleteTeamTask(taskID, activeTeam.ID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	err = m.deps.taskRepo.DeleteNeedAttentionTask(taskID, userID, activeTeam.ID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	return true, nil
}

func (m Mutation) CompleteTask(ctx context.Context, args struct {
	TaskID graphql.ID
}) (bool, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return false, err
	}

	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	activeTeam, err := m.deps.teamRepo.FindActiveTeam(userID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	needAttentionTask, err := m.deps.taskRepo.FindTaskNeedAttentionForUser(userID, activeTeam.ID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	if needAttentionTask.ID != taskID {
		return true, err
	}

	_, err = m.deps.taskRepo.SetNeedAttentionTask(nil, userID, activeTeam.ID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	err = m.deps.taskRepo.SetTeamTaskStatus(taskID, activeTeam.ID, entity.TaskStatusDelivered)
	if err != nil {
		log.Println(err)
		return false, err
	}

	return true, nil
}

func (m Mutation) UpdateTask(ctx context.Context, args struct {
	TaskID graphql.ID
	Task   TaskInput
}) bool {
	panic("not implemented")
}

func NewMutation(deps *Dependencies, prototypeDeps *resolver.Dependencies, query *Query) Mutation {
	return Mutation{
		deps:          deps,
		prototypeDeps: prototypeDeps,
		query:         query,
	}
}
