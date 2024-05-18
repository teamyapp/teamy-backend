package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AttachmentList interface {
	FindAttachmentListByIDWithTx(
	    ct context.Context, 
	    tx *transaction.Transaction, 
	    attachmentListID uint64,
	) (entity.AttachmentList, *errs.Error)
	FindAttachmentListsByOwnerIDAndOwnerTypeWithTx(
	    ct context.Context, 
	    tx *transaction.Transaction, 
	    ownerType entity.AttachmentListOwnerType,
	    ownerID uint64,
	) ([]entity.AttachmentList, *errs.Error)
	FindAttachmentListsWithTx(
	    ct context.Context, 
	    tx *transaction.Transaction,
	    ownerType entity.AttachmentListOwnerType, 
	    ownerID uint64, 
	    listLabel string
	 ) ([]entity.AttachmentList, *errs.Error)
	CreateAttachmentList(ct context.Context, tx *transaction.Transaction, attachmentList entity.AttachmentList) *errs.Error
}
