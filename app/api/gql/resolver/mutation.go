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
	data        *resolver.Data
	taskService service.Task
}

func (m Mutation) CreateTask(ctx context.Context, args struct {
	Task TaskInput
}) (bool, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return false, err
	}

	task, err := fromGraphQLTaskInput(args.Task)
	if err != nil {
		log.Println(err)
		return false, err
	}

	taskID, err := m.taskService.CreateTask(task, userID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	err = m.data.CreateLifetimeEvent(graphql.ID(fmt.Sprint(userID)), resolver.LifetimeEventType{
		Type: resolver.Creation,
		Creation: &resolver.EventCreation{
			TaskID: graphql.ID(fmt.Sprint(taskID)),
		},
	})
	if err != nil {
		log.Println(err)
	}

	return true, nil
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
