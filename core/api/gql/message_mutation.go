package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/service"
)

func (m Mutation) CreateMessage(ct context.Context, args struct {
	ThreadID graphql.ID
	Message  struct {
		Body string
	}
}) (Message, error) {
	threadID, argErr := fromGraphQLID(args.ThreadID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Message{}, errs.ToResolverErr(internalErr)
	}

	input := service.CreateMessageInput{Body: args.Message.Body}
	message, err := m.deps.threadService.CreateMessage(ct, threadID, input)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Message{}, errs.ToResolverErr(err)
	}

	return newMessage(m.deps, message), nil
}

func (m Mutation) UpdateMessage(ct context.Context, args struct {
	MessageID graphql.ID
	Input     struct {
		Body string
	}
}) (Message, error) {
	messageID, argErr := fromGraphQLID(args.MessageID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Message{}, errs.ToResolverErr(internalErr)
	}

	input := service.UpdateMessageInput{Body: args.Input.Body}
	message, err := m.deps.threadService.UpdateMessage(ct, messageID, input)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Message{}, errs.ToResolverErr(err)
	}

	return newMessage(m.deps, message), nil
}

func (m Mutation) DeleteMessage(ct context.Context, args struct {
	MessageID graphql.ID
}) (Message, error) {
	messageID, argErr := fromGraphQLID(args.MessageID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Message{}, errs.ToResolverErr(internalErr)
	}

	message, err := m.deps.threadService.DeleteMessage(ct, messageID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Message{}, errs.ToResolverErr(err)
	}

	return newMessage(m.deps, message), nil
}
