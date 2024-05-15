package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Image interface {
	FindImagesByAttachmentListIDWithTx(ct context.Context, tx *transaction.Transaction, attachmentListID uint64) ([]entity.Image, *errs.Error)
	CreateImage(ct context.Context, tx *transaction.Transaction, image entity.Image) *errs.Error
}
