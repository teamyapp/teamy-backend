package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskLink struct {
	deps     *Dependencies
	taskLink entity.TaskLink
}

func (t TaskLink) ID() graphql.ID {
	return toGraphQLID(t.taskLink.ID)
}

func (t TaskLink) TaskID() graphql.ID {
	return toGraphQLID(t.taskLink.TaskID)
}

func (t TaskLink) Title() string {
	return t.taskLink.Title
}

func (t TaskLink) Url() string {
	return t.taskLink.URL
}

func (t TaskLink) IconUrl() *string {
	return t.taskLink.IconURL
}

func (t TaskLink) IconHoverUrl() *string {
	return t.taskLink.IconHoverURL
}

func (t TaskLink) CreatedAt(ct context.Context) graphql.Time {
	return toGraphQLTime(t.taskLink.CreatedAt)
}

func (t TaskLink) UpdatedAt(ct context.Context) *graphql.Time {
	return toGraphQLTimePtr(t.taskLink.UpdatedAt)
}

func newTaskLink(deps *Dependencies, taskLink entity.TaskLink) TaskLink {
	return TaskLink{
		deps:     deps,
		taskLink: taskLink,
	}
}
