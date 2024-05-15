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

func (a AttachmentList) Images(ct context.Context) ([]Image, error) {
	images, err := a.deps.attachmentService.FindImagesByAttachmentListID(ct, a.attachmentList.ListID)
	if err != nil {
		a.deps.logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(images, func(image entity.Image, _ int) Image {
		return newImage(a.deps, image)
	}), nil
}

func newAttachmentList(deps *Dependencies, attachmentList entity.AttachmentList) AttachmentList {
	return AttachmentList{
		deps:           deps,
		attachmentList: attachmentList,
	}
}
