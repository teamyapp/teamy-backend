package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Image struct {
	deps  *Dependencies
	image entity.Image
}

func (i Image) ID() graphql.ID {
	return toGraphQLID(i.image.ID)
}

func (i Image) URL() string {
	return i.image.URL
}

func (i Image) Size() int32 {
	return int32(i.image.Size)
}

func (i Image) CreatedAt() graphql.Time {
	return toGraphQLTime(i.image.CreatedAt)
}

func (i Image) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(i.image.UpdatedAt)
}

func (i Image) AttachmentList(ct context.Context) AttachmentList {
	attachmentList, err := i.deps.attachmentService.FindAttachmentListByID(ct, i.image.AttachmentListID)
	if err != nil {
		i.deps.logger.ErrorWithContext(ct, err)
		return AttachmentList{}
	}

	return newAttachmentList(i.deps, attachmentList)
}

func newImage(deps *Dependencies,
	image entity.Image) Image {
	return Image{
		deps:  deps,
		image: image,
	}
}
