package resolver

import (
	"context"
	"log"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/identity"
	"github.com/teamyapp/teamy-backend/app/service"
)

type Query struct {
	executionService service.Execution
}

func (q Query) ExecutionMode(ctx context.Context) (ExecutionMode, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return ExecutionMode{}, err
	}
	return ExecutionMode{
		userID:           userID,
		executionService: q.executionService,
	}, nil
}

func (q Query) Task(ctx context.Context, args struct {
	TaskID graphql.ID
}) *Task {
	panic("not implemented")
}

func NewQuery(executionService service.Execution) Query {
	return Query{executionService: executionService}
}
