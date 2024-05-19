package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/service"
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

func (t TaskLink) PreviewTitle() string {
	return t.taskLink.PreviewTitle
}

func (t TaskLink) URL() string {
	return t.taskLink.URL
}

func (t TaskLink) IconURL() *string {
	return t.taskLink.IconURL
}

func (t TaskLink) IconHoverURL() *string {
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

func (m Mutation) CreateTaskLink(
	ctx context.Context,
	args struct {
		TaskID graphql.ID
		Input  struct {
			Title        string
			PreviewTitle string
			URL          string
			IconURL      *string
			IconHoverURL *string
		}
	},
) (TaskLink, error) {
	taskID, argErr := fromGraphQLID(args.TaskID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return TaskLink{}, errs.ToResolverErr(internalErr)
	}

	taskLink := service.CreateTaskLinkInput{
		TaskID:       taskID,
		Title:        args.Input.Title,
		PreviewTitle: args.Input.PreviewTitle,
		URL:          args.Input.URL,
		IconURL:      args.Input.IconURL,
		IconHoverURL: args.Input.IconHoverURL,
	}
	taskLinkRes, err := m.deps.taskLinkService.CreateTaskLink(
		ctx,
		taskLink,
	)

	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return TaskLink{}, errs.ToResolverErr(err)
	}

	return newTaskLink(m.deps, taskLinkRes), nil
}

func (m Mutation) DeleteTaskLink(
	ctx context.Context,
	args struct {
		ID graphql.ID
	},
) (TaskLink, error) {
	id, argErr := fromGraphQLID(args.ID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return TaskLink{}, errs.ToResolverErr(internalErr)
	}

	task, err := m.deps.taskLinkService.DeleteTaskLink(ctx, id)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return TaskLink{}, errs.ToResolverErr(err)
	}

	return newTaskLink(m.deps, task), nil
}
