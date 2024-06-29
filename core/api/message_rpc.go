package api

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
	pbteamy "github.com/teamyapp/protocol/pb/pbgo/teamy"
	pbmessage "github.com/teamyapp/protocol/pb/pbgo/teamy/message"
	"github.com/teamyapp/teamy-backend/core/service"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MessageRPC struct {
	logger        telemetry.Logger
	threadService service.Thread
	pbteamy.UnimplementedMessageServiceServer
}

var _ pbteamy.MessageServiceServer = (*MessageRPC)(nil)
var _ runner.Service = (*MessageRPC)(nil)

func (m MessageRPC) Start(runner *runner.ServiceRunner) *errs.Error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		pbteamy.RegisterMessageServiceServer(server, m)
	})
	return nil
}

func (m MessageRPC) ListMessages(ct context.Context, req *pbteamy.ListMessagesRequest) (*pbteamy.ListMessagesResponse, error) {
	messages, err := m.threadService.FindMessages(ct, req.ThreadId)
	if err != nil {
		m.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	var pbMessages []*pbmessage.Message
	for _, message := range messages {
		pbMessages = append(pbMessages, &pbmessage.Message{
			Id:           message.ID,
			ThreadId:     message.ThreadID,
			AuthorUserId: message.AuthorUserID,
			Body:         message.Body,
			CreatedAt:    toProtoTimePtr(&message.CreatedAt),
			UpdatedAt:    toProtoTimePtr(message.UpdatedAt),
		})
	}

	return &pbteamy.ListMessagesResponse{
		Messages: pbMessages,
	}, nil
}

func (m MessageRPC) CreateMessage(ct context.Context, req *pbteamy.CreateMessageRequest) (*pbteamy.CreateMessageResponse, error) {
	createMessageInput := service.CreateMessageInput{
		Body: req.Body,
	}

	message, err := m.threadService.CreateMessage(ct, req.ThreadId, createMessageInput)
	if err != nil {
		m.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &pbteamy.CreateMessageResponse{
		Message: &pbmessage.Message{
			Id:           message.ID,
			ThreadId:     message.ThreadID,
			AuthorUserId: message.AuthorUserID,
			Body:         message.Body,
			CreatedAt:    toProtoTimePtr(&message.CreatedAt),
			UpdatedAt:    toProtoTimePtr(message.UpdatedAt),
		},
	}, nil
}

func (m MessageRPC) UpdateMessage(ct context.Context, req *pbteamy.UpdateMessageRequest) (*pbteamy.UpdateMessageResponse, error) {
	updateMessageInput := service.UpdateMessageInput{
		Body: req.Body,
	}

	message, err := m.threadService.UpdateMessage(ct, req.Id, updateMessageInput)
	if err != nil {
		m.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &pbteamy.UpdateMessageResponse{
		Message: &pbmessage.Message{
			Id:           message.ID,
			ThreadId:     message.ThreadID,
			AuthorUserId: message.AuthorUserID,
			Body:         message.Body,
			CreatedAt:    toProtoTimePtr(&message.CreatedAt),
			UpdatedAt:    toProtoTimePtr(message.UpdatedAt),
		},
	}, nil
}

func (m MessageRPC) DeleteMessage(ct context.Context, req *pbteamy.DeleteMessageRequest) (*emptypb.Empty, error) {
	_, err := m.threadService.DeleteMessage(ct, req.Id)
	if err != nil {
		m.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &emptypb.Empty{}, nil
}

func NewMessageRPC(logger telemetry.Logger, threadService service.Thread) MessageRPC {
	return MessageRPC{
		logger:        logger,
		threadService: threadService,
	}
}
