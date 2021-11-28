package resolver

import (
	"context"
	"fmt"
	"log"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/one/identity"
	"github.com/teamyapp/teamy-backend/app/api/gqlv2/resolver"
	"github.com/teamyapp/teamy-backend/app/service"
)

type Mutation struct {
	dep         *resolver.Dependencies
	taskService service.Task
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

	taskID, err := m.taskService.CreateTask(task, userID)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	err = m.dep.Data.CreateLifetimeEvent(graphql.ID(fmt.Sprint(userID)), resolver.LifetimeEventType{
		Type: resolver.Creation,
		Creation: &resolver.EventCreation{
			TaskID: graphql.ID(fmt.Sprint(taskID)),
		},
	})
	if err != nil {
		log.Println(err)
	}

	task, err = m.taskService.FindTask(taskID)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	taskResolver := newTask(task)
	taskResolver.dep = m.dep

	m.dep.Data.CreationRelations = append(m.dep.Data.CreationRelations, resolver.CreationRelation{
		TaskID: taskResolver.ID(),
		UserID: graphql.ID(fmt.Sprint(userID)),
	})

	return taskResolver, nil
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

	err = m.taskService.StartTask(taskID, userID)
	if err != nil {
		log.Println(err)
		return false, err
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

	err = m.taskService.DeleteTask(taskID, userID)
	if err != nil {
		log.Println(err)
	}

	return true, err
}

func (m Mutation) UpdateTask(ctx context.Context, args struct {
	TaskID graphql.ID
	Task   TaskInput
}) bool {
	panic("not implemented")
}

func (m Mutation) PerformTaskAction(ctx context.Context, args struct {
	TaskID graphql.ID
	Action TaskAction
}) bool {
	panic("not implemented")
}

func NewMutation(taskService service.Task) Mutation {
	return Mutation{
		taskService: taskService,
	}
}
