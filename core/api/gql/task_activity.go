package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskActivity struct {
	taskActivity entity.TaskActivity
}

func (t TaskActivity) TaskID(ct context.Context) graphql.ID {
	return toGraphQLID(t.taskActivity.TaskID)
}

func (t TaskActivity) DragTaskActivity(ct context.Context) DragTaskActivity {
	return newDragTaskActivity(t.taskActivity.DragTaskActivity)
}

func newTaskActivity(taskActivity entity.TaskActivity) TaskActivity {
	return TaskActivity{
		taskActivity: taskActivity,
	}
}
