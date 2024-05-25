package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AttachmentList struct {
	deps           *Dependencies
	attachmentList entity.AttachmentList
}

func (a AttachmentList) ListID() graphql.ID {
	return toGraphQLID(a.attachmentList.ListID)
}

func (a AttachmentList) OwnerType() entity.AttachmentListOwnerType {
	return a.attachmentList.OwnerType
}

func (a AttachmentList) OwnerID() graphql.ID {
	return toGraphQLID(a.attachmentList.OwnerID)
}

func (a AttachmentList) ListLabel() string {
	return a.attachmentList.ListLabel
}

func (a AttachmentList) CreatedAt() graphql.Time {
	return toGraphQLTime(a.attachmentList.CreatedAt)
}

func (a AttachmentList) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(a.attachmentList.UpdatedAt)
}

func (a AttachmentList) Attachments(ct context.Context) ([]Attachment, error) {
	attachments, err := a.deps.attachmentService.FindAttachmentsByAttachmentListID(ct, a.attachmentList.ListID)
	if err != nil {
		a.deps.logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(attachments, func(attachment entity.Attachment, _ int) Attachment {
		return newAttachment(a.deps, attachment)
	}), nil
}

func newAttachmentList(deps *Dependencies, attachmentList entity.AttachmentList) AttachmentList {
	return AttachmentList{
		deps:           deps,
		attachmentList: attachmentList,
	}
}

func (m Mutation) CreateAttachmentListFileUploadSession(
	ct context.Context,
	args struct {
		AttachmentListID graphql.ID
	},
) (graphql.ID, error) {
	attachmentListID, argErr := fromGraphQLID(args.AttachmentListID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return "", errs.ToResolverErr(internalErr)
	}

	sessionID, err := m.deps.attachmentService.CreateAttachmentListFileUploadSession(ct, attachmentListID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return "", errs.ToResolverErr(err)
	}

	return toGraphQLID(sessionID), nil
}

func (m Mutation) FinishAttachmentListFileUploadSession(ct context.Context, args struct {
	AttachmentListID    graphql.ID
	FileUploadSessionID graphql.ID
}) (Attachment, error) {
	attachmentListID, argErr := fromGraphQLID(args.AttachmentListID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Attachment{}, errs.ToResolverErr(internalErr)
	}

	fileUploadSessionID, argErr := fromGraphQLID(args.FileUploadSessionID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Attachment{}, errs.ToResolverErr(internalErr)
	}

	attachment, err := m.deps.attachmentService.FinishAttachmentListFileUploadSession(ct, attachmentListID, fileUploadSessionID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Attachment{}, errs.ToResolverErr(err)
	}

	return newAttachment(m.deps, attachment), nil
}

func (m Mutation) DeleteAttachmentListFile(ct context.Context, args struct {
	AttachmentID graphql.ID
}) (Attachment, error) {
	attachmentID, argErr := fromGraphQLID(args.AttachmentID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Attachment{}, errs.ToResolverErr(internalErr)
	}

	attachment, err := m.deps.attachmentService.DeleteAttachment(ct, attachmentID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Attachment{}, errs.ToResolverErr(err)
	}

	return newAttachment(m.deps, attachment), nil
}
