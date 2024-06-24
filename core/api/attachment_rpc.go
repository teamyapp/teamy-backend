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
)

type AttachmentRPC struct {
	logger            telemetry.Logger
	attachmentService *service.Attachment
	pbteamy.UnimplementedAttachmentServiceServer
}

var _ pbteamy.AttachmentServiceServer = (*AttachmentRPC)(nil)
var _ runner.Service = (*AttachmentRPC)(nil)

func (a AttachmentRPC) Start(runner *runner.ServiceRunner) *errs.Error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		pbteamy.RegisterAttachmentServiceServer(server, a)
	})
	return nil
}

func (a AttachmentRPC) GetAttachmentList(ct context.Context, req *pbteamy.GetAttachmentListRequest) (*pbteamy.GetAttachmentListResponse, error) {
	attachmentList, err := a.attachmentService.FindAttachmentList(ct, protoAttachmentListOwnerTypes[req.OwnerType], req.OwnerId, req.ListLabel)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &pbteamy.GetAttachmentListResponse{
		AttachmentList: &pbmessage.AttachmentList{
			OwnerId:   attachmentList.OwnerID,
			ListLabel: attachmentList.ListLabel,
			ListId:    attachmentList.ListID,
			OwnerType: attachmentListOwnerTypes[attachmentList.OwnerType],
			CreatedAt: toProtoTimePtr(&attachmentList.CreatedAt),
			UpdatedAt: toProtoTimePtr(attachmentList.UpdatedAt),
		},
	}, nil
}

func (a AttachmentRPC) ListAttachments(ct context.Context, req *pbteamy.ListAttachmentsRequest) (*pbteamy.ListAttachmentsResponse, error) {
	attachments, err := a.attachmentService.FindAttachmentsByAttachmentListID(ct, req.AttachmentListId)
	if err != nil {
		a.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	var pbAttachments []*pbmessage.Attachment
	for _, attachment := range attachments {
		pbAttachments = append(pbAttachments, &pbmessage.Attachment{
			Id:               attachment.ID,
			AttachmentListId: attachment.AttachmentListID,
			Type:             attachmentTypes[attachment.Type],
			Url:              attachment.URL,
			Size:             attachment.Size,
			CreatedAt:        toProtoTimePtr(&attachment.CreatedAt),
			UpdatedAt:        toProtoTimePtr(attachment.UpdatedAt),
		})
	}

	return &pbteamy.ListAttachmentsResponse{
		Attachments: pbAttachments,
	}, nil
}

func NewAttachmentRPC(logger telemetry.Logger, attachmentService *service.Attachment) AttachmentRPC {
	return AttachmentRPC{
		logger:            logger,
		attachmentService: attachmentService,
	}
}
