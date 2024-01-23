package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Tag struct {
	deps *Dependencies
	tag  entity.Tag
}

func (t Tag) ID(ctx context.Context) graphql.ID {
	return toGraphQLID(t.tag.ID)
}

func (t Tag) Tag(ctx context.Context) string {
	return t.tag.Tag
}

func (m Mutation) AddTagToApp(
	ctx context.Context,
	args struct {
		AppID graphql.ID
		Tag   string
	},
) (Tag, error) {
	appID, internalErr := fromGraphQLID(args.AppID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return Tag{}, errs.ToResolverErr(internalErr)
	}

	tag, err := m.deps.appService.AddTagToApp(ctx, appID, args.Tag)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return Tag{}, errs.ToResolverErr(err)
	}

	return newTag(m.deps, tag), nil
}

func (m Mutation) RemoveTagFromApp(
	ctx context.Context,
	args struct {
		AppID graphql.ID
		TagID graphql.ID
	},
) (Tag, error) {
	appID, internalErr := fromGraphQLID(args.AppID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return Tag{}, errs.ToResolverErr(internalErr)
	}

	tagID, internalErr := fromGraphQLID(args.TagID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return Tag{}, errs.ToResolverErr(internalErr)
	}

	tag, err := m.deps.appService.RemoveTagFromApp(ctx, appID, tagID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return Tag{}, errs.ToResolverErr(err)
	}

	return newTag(m.deps, tag), nil
}

func newTag(deps *Dependencies, tag entity.Tag) Tag {
	return Tag{
		deps: deps,
		tag:  tag,
	}
}
