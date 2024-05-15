package service

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	cloudTransaction "github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/transaction"
)

type Attachment struct {
	transactionGroupFactory transaction.GroupFactory
	attachmentListDao       dao.AttachmentList
	imageDao                dao.Image
}

func (a *Attachment) FindAttachmentListByID(ct context.Context, listID uint64) (entity.AttachmentList, *errs.Error) {
	var attachmentList entity.AttachmentList
	err := a.transactionGroupFactory.WithTransactionGroup(ct, true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		attachmentList, internalErr = a.attachmentListDao.FindAttachmentListByIDWithTx(ct, tx, listID)
		return internalErr
	})

	return attachmentList, err
}

func (a *Attachment) FindAttachmentList(ct context.Context, ownerID uint64, ownerType entity.AttachmentListOwnerType, listLabel string) (*entity.AttachmentList, *errs.Error) {
	var attachmentList *entity.AttachmentList
	err := a.transactionGroupFactory.WithTransactionGroup(ct, true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		attachmentLists, internalErr := a.attachmentListDao.FindAttachmentListsWithTx(ct, tx, ownerID, ownerType, listLabel)
		if internalErr != nil {
			return internalErr
		}

		if len(attachmentLists) == 0 {
			return nil
		}

		if len(attachmentLists) > 1 {
			return errs.NewError(errs.Unknown, "Found multiple attachment lists")
		}

		attachmentList = &attachmentLists[0]
		return internalErr
	})

	return attachmentList, err
}

func (a *Attachment) FindImagesByAttachmentListID(ct context.Context, attachmentListID uint64) ([]entity.Image, *errs.Error) {
	var images []entity.Image
	err := a.transactionGroupFactory.WithTransactionGroup(ct, true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		images, internalErr = a.imageDao.FindImagesByAttachmentListIDWithTx(ct, tx, attachmentListID)
		return internalErr
	})

	return images, err
}

func NewAttachment(transactionGroupFactory transaction.GroupFactory, attachmentListDao dao.AttachmentList, imageDao dao.Image) *Attachment {
	return &Attachment{
		transactionGroupFactory: transactionGroupFactory,
		attachmentListDao:       attachmentListDao,
		imageDao:                imageDao,
	}
}
