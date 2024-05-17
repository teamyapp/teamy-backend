package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Attachment struct {
	deps       *Dependencies
	attachment entity.Attachment
}

func (a Attachment) ID() graphql.ID {
	return toGraphQLID(a.attachment.ID)
}

func (a Attachment) Type() entity.AttachmentType {
	return a.attachment.Type
}

func (i Attachment) URL() string {
	return i.attachment.URL
}

func (i Attachment) Size() int32 {
	return int32(i.attachment.Size)
}

func (i Attachment) CreatedAt() graphql.Time {
	return toGraphQLTime(i.attachment.CreatedAt)
}

func (i Attachment) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(i.attachment.UpdatedAt)
}

func (i Attachment) AttachmentList(ct context.Context) AttachmentList {
	attachmentList, err := i.deps.attachmentService.FindAttachmentListByID(ct, i.attachment.AttachmentListID)
	if err != nil {
		i.deps.logger.ErrorWithContext(ct, err)
		return AttachmentList{}
	}

	return newAttachmentList(i.deps, attachmentList)
}

func newAttachment(deps *Dependencies,
	attachment entity.Attachment) Attachment {
	return Attachment{
		deps:       deps,
		attachment: attachment,
	}
}
