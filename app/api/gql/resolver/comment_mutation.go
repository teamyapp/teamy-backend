package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/teamyapp/one/identity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

func (m Mutation) CreateComment(
	ctx context.Context,
	args struct {
		TaskID  graphql.ID
		Content string
	},
) (Comment, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		return Comment{}, err
	}
	c, err := m.deps.Data.CreateComment(entity.Comment{
		Content:     args.Content,
		CommenterID: toGraphQLID(userID),
		TaskID:      args.TaskID,
	})
	if err != nil {
		return Comment{}, err
	}
	return Comment{
		deps:    m.deps,
		Comment: c,
	}, nil
}

func (m Mutation) Comment(
	ctx context.Context,
	args struct {
		ID graphql.ID
	},
) (CommentUpdate, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		return CommentUpdate{}, err
	}
	comments := m.deps.Data.FilterComments(func(c entity.Comment) bool {
		return toGraphQLID(c.ID) == args.ID
	})
	if len(comments) == 0 {
		return CommentUpdate{}, errors.Errorf("comment %v is not found", args.ID)
	}
	if comments[0].CommenterID != toGraphQLID(userID) {
		return CommentUpdate{}, errors.Errorf("user %v is not the commenter of comment %v", userID, args.ID)
	}
	return CommentUpdate{
		deps:    m.deps,
		comment: comments[0],
	}, nil
}

type CommentUpdate struct {
	deps    *Dependencies
	comment entity.Comment
}

func (cu CommentUpdate) UpdateContent(
	ctx context.Context,
	args struct {
		Content string
	},
) (CommentUpdate, error) {
	newComment, err := cu.deps.Data.UpdateComment(cu.comment.ID, func(c entity.Comment) entity.Comment {
		c.Content = args.Content
		return c
	})
	if err != nil {
		return cu, err
	}
	cu.comment = newComment
	return cu, nil
}

func (cu CommentUpdate) Comment() Comment {
	return Comment{
		deps:    cu.deps,
		Comment: cu.comment,
	}
}
