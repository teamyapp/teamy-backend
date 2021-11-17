package resolver

import (
	"context"
	"log"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/one/identity"
	"github.com/teamyapp/teamy-backend/app/service"
)

type Mutation struct {
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

	return true, m.taskService.CreateTask(task, userID)
}

func (m Mutation) DeleteTask(ctx context.Context, args struct {
	TaskID graphql.ID
}) bool {
	panic("not implemented")
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
