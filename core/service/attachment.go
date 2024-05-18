package service

import (
	"context"
	"fmt"
	"time"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudTransaction "github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/transaction"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Attachment struct {
	cloudClientRegistry            *client.Registry
	transactionGroupFactory        transaction.GroupFactory
	cloudWebAPIExternalBaseURL     string
	logger                         telemetry.Logger
	stateSyncer                    *realtime.StateSyncer
	attachmentListDao              dao.AttachmentList
	attachmentDao                  dao.Attachment
	attachmentFileUploadSessionDao dao.AttachmentFileUploadSession
	taskDao                        dao.Task
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

func (a *Attachment) FindAttachmentsByAttachmentListID(ct context.Context, attachmentListID uint64) ([]entity.Attachment, *errs.Error) {
	var attachments []entity.Attachment
	err := a.transactionGroupFactory.WithTransactionGroup(ct, true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		attachments, internalErr = a.attachmentDao.FindAttachmentsByAttachmentListIDWithTx(ct, tx, attachmentListID)
		return internalErr
	})

	return attachments, err
}

func (a *Attachment) CreateAttachmentListFileUploadSession(ct context.Context, attachmentListID uint64) (uint64, *errs.Error) {
	res, rpcErr := a.cloudClientRegistry.FileClient().CreateUploadSession(ct, &emptypb.Empty{})
	if rpcErr != nil {
		return 0, errs.FromGRPCErr(rpcErr)
	}

	fileUploadSession := entity.AttachmentFileUploadSession{
		AttachmentListID:    attachmentListID,
		FileUploadSessionID: res.UploadSessionId,
		IsCompleted:         false,
		CreatedAt:           time.Now(),
	}
	err := a.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			return a.attachmentFileUploadSessionDao.CreateAttachmentFileUploadSession(ct, tx, fileUploadSession)
		})

	return res.UploadSessionId, err
}

func (a *Attachment) FinishAttachmentListFileUploadSession(
	ct context.Context,
	attachmentListID uint64,
	fileUploadSessionID uint64,
) (entity.Attachment, *errs.Error) {
	findUploadSessionReq := proto.FindUploadSessionRequest{
		UploadSessionId: fileUploadSessionID,
	}
	uploadSession, rpcErr := a.cloudClientRegistry.FileClient().FindUploadSession(ct, &findUploadSessionReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.Attachment{}, internalErr
	}

	var attachmentFileUploadSession entity.AttachmentFileUploadSession
	var attachment entity.Attachment
	err := a.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var internalErr *errs.Error
			attachmentFileUploadSession, internalErr = a.attachmentFileUploadSessionDao.FindAttachmentFileUploadSessionWithTx(
				ct,
				tx,
				attachmentListID,
				fileUploadSessionID)
			if internalErr != nil {
				return internalErr
			}

			if attachmentFileUploadSession.IsCompleted {
				return errs.NewError(errs.InvalidOperation, fmt.Sprintf("attachment file upload session is already completed: attachmentListID=%v, fileUploadSessionID=%v",
					attachmentListID,
					fileUploadSessionID))
			}

			now := time.Now().UTC()
			attachmentFileUploadSession.IsCompleted = true
			attachmentFileUploadSession.UpdatedAt = &now
			internalErr = a.attachmentFileUploadSessionDao.UpdateAttachmentFileUploadSession(ct, tx, attachmentFileUploadSession)
			if internalErr != nil {
				return internalErr
			}

			attachmentURL := io.GetFileURL(a.cloudWebAPIExternalBaseURL, uploadSession.FileId)

			var attachmentType entity.AttachmentType
			switch uploadSession.MimeType {
			case "image/jpeg", "image/png", "image/gif":
				attachmentType = entity.AttachmentTypeImage
			default:
				return errs.NewError(errs.InvalidOperation, fmt.Sprintf("unsupported attachment type: %v", uploadSession.MimeType))
			}

			attachment = entity.Attachment{
				ID:               uploadSession.FileId,
				Type:             attachmentType,
				URL:              attachmentURL,
				Size:             uploadSession.TotalSizeInBytes,
				AttachmentListID: attachmentListID,
				CreatedAt:        now,
			}

			createImageMutation := mutation.NewCreateAttachment(
				a.logger,
				a.stateSyncer,
				a.attachmentListDao,
				a.attachmentDao,
				a.taskDao,
				attachment,
			)

			internalErr = createImageMutation.Execute(ct, tx)
			if internalErr != nil {
				return internalErr
			}

			rtTx.AppendMutation(createImageMutation)
			return nil
		})

	return attachment, err
}

func NewAttachment(
	cloudClientRegistry *client.Registry,
	transactionGroupFactory transaction.GroupFactory,
	cloudWebAPIExternalBaseURL string,
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	attachmentListDao dao.AttachmentList,
	attachmentDao dao.Attachment,
	attachmentFileUploadSessionDao dao.AttachmentFileUploadSession,
	taskDao dao.Task,
) *Attachment {
	return &Attachment{
		cloudClientRegistry:            cloudClientRegistry,
		transactionGroupFactory:        transactionGroupFactory,
		cloudWebAPIExternalBaseURL:     cloudWebAPIExternalBaseURL,
		logger:                         logger,
		stateSyncer:                    stateSyncer,
		attachmentListDao:              attachmentListDao,
		attachmentDao:                  attachmentDao,
		attachmentFileUploadSessionDao: attachmentFileUploadSessionDao,
		taskDao:                        taskDao,
	}
}
