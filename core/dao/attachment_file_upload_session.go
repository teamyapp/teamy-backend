package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AttachmentFileUploadSession interface {
	FindAttachmentFileUploadSessionWithTx(
		ct context.Context,
		tx *transaction.Transaction,
		attachmentListID uint64,
		fileUploadSessionID uint64,
	) (entity.AttachmentFileUploadSession, *errs.Error)
	CreateAttachmentFileUploadSession(ct context.Context, tx *transaction.Transaction, attachmentFileUploadSession entity.AttachmentFileUploadSession) *errs.Error
	UpdateAttachmentFileUploadSession(ct context.Context, tx *transaction.Transaction, attachmentFileUploadSession entity.AttachmentFileUploadSession) *errs.Error
}
