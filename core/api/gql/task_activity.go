package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskActivity struct {
	deps         *Dependencies
	taskActivity entity.TeamTaskDraggingActivity
}

func (t TaskActivity) TaskID(ct context.Context) graphql.ID {
	return toGraphQLID(t.taskActivity.TaskID)
}

func (t TaskActivity) TeamID(ct context.Context) graphql.ID {
	return toGraphQLID(t.taskActivity.TeamID)
}

func (t TaskActivity) IsDragging(ct context.Context) bool {
	return t.taskActivity.IsDragging
}

func (t TaskActivity) DragByUserID(ct context.Context) graphql.ID {
	return toGraphQLID(t.taskActivity.DragByUserID)
}

func (t TaskActivity) DraggingClientID(ct context.Context) graphql.ID {
	return toGraphQLID(t.taskActivity.DraggingClientID)
}

func newTaskActivity(deps *Dependencies, taskActivity entity.TeamTaskDraggingActivity) TaskActivity {
	return TaskActivity{
		deps:         deps,
		taskActivity: taskActivity,
	}
}
