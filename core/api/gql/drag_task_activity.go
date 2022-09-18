package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type DragTaskActivity struct {
	dragTaskActivity entity.DragTaskActivity
}

func (d DragTaskActivity) IsDragging(ct context.Context) bool {
	return d.dragTaskActivity.IsDragging
}

func (d DragTaskActivity) DragByUserID(ct context.Context) graphql.ID {
	return toGraphQLID(d.dragTaskActivity.DragByUserID)
}

func (d DragTaskActivity) DraggingClientID(ct context.Context) graphql.ID {
	return toGraphQLID(d.dragTaskActivity.DraggingClientID)
}

func newDragTaskActivity(dragTaskActivity entity.DragTaskActivity) DragTaskActivity {
	return DragTaskActivity{
		dragTaskActivity: dragTaskActivity,
	}
}
