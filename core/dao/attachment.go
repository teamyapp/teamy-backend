package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Attachment interface {
	FindAttachmentsByAttachmentListIDWithTx(ct context.Context, tx *transaction.Transaction, attachmentListID uint64) ([]entity.Attachment, *errs.Error)
	CreateAttachment(ct context.Context, tx *transaction.Transaction, attachment entity.Attachment) *errs.Error
}
