package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AttachmentList interface {
	FindAttachmentListByIDWithTx(ct context.Context, tx *transaction.Transaction, attachmentListID uint64) (entity.AttachmentList, *errs.Error)
	FindAttachmentListsByOwnerIDAndOwnerTypeWithTx(ct context.Context, tx *transaction.Transaction, ownerID uint64, ownerType entity.AttachmentListOwnerType) ([]entity.AttachmentList, *errs.Error)
	FindAttachmentListsWithTx(ct context.Context, tx *transaction.Transaction, ownerID uint64, ownerType entity.AttachmentListOwnerType, ListLabel string) ([]entity.AttachmentList, *errs.Error)
	CreateAttachmentList(ct context.Context, tx *transaction.Transaction, attachmentList entity.AttachmentList) *errs.Error
}
