package gql

import (
	"context"

	"github.com/teamyapp/teamy-backend/core/entity"
)

type DragTaskActivity struct {
	deps             *Dependencies
	dragTaskActivity entity.DragTaskActivity
}

func (d DragTaskActivity) IsDragging(ct context.Context) bool {
	return d.dragTaskActivity.IsDragging
}

func (d DragTaskActivity) Client(ct context.Context) Client {
	return newClient(d.deps, d.dragTaskActivity.Client)
}

func newDragTaskActivity(deps *Dependencies, dragTaskActivity entity.DragTaskActivity) DragTaskActivity {
	return DragTaskActivity{
		deps:             deps,
		dragTaskActivity: dragTaskActivity,
	}
}
