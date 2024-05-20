package service

import (
	"context"
	"time"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudTransaction "github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/transaction"
)

type Backfill struct {
	logger                  telemetry.Logger
	transactionGroupFactory transaction.GroupFactory
	cloudClientRegistry     *client.Registry
	taskDao                 dao.Task
	attachmentListDao       dao.AttachmentList
}

func (b *Backfill) backfillAttachmentListForTaskContext(ct context.Context) *errs.Error {
	return b.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		tasks, internalErr := b.taskDao.FindAllTasksWithTx(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		for _, task := range tasks {
			attachmentLists, internalErr := b.attachmentListDao.FindAttachmentListsWithTx(ct, tx, entity.AttachmentListOwnerTypeTask, task.ID, "context")
			if internalErr != nil {
				return internalErr
			}

			if len(attachmentLists) > 0 {
			   continue;
			}
				genAttachmentListIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "attachmentListID"}
				genAttachmentListIDRes, rpcErr := b.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genAttachmentListIDReq)
				if rpcErr != nil {
					return errs.FromGRPCErr(rpcErr)
				}

				attachmentList := entity.AttachmentList{
					OwnerType: entity.AttachmentListOwnerTypeTask,
					OwnerID:   task.ID,
					ListLabel: "context",
					ListID:    genAttachmentListIDRes.UniqueNumber,
					CreatedAt: time.Now().UTC(),
				}

				return b.attachmentListDao.CreateAttachmentList(ct, tx, attachmentList)
		}

		return nil
	})
}

func (b *Backfill) BackfillData(ct context.Context) *errs.Error {
	return b.addAttachmentListToTaskContext(ct)
}

func NewBackfill(
	logger telemetry.Logger,
	transactionGroupFactory transaction.GroupFactory,
	cloudClientRegistry *client.Registry,
	taskDao dao.Task,
	attachmentListDao dao.AttachmentList,
) *Backfill {
	return &Backfill{
		logger:                  logger,
		transactionGroupFactory: transactionGroupFactory,
		cloudClientRegistry:     cloudClientRegistry,
		taskDao:                 taskDao,
		attachmentListDao:       attachmentListDao,
	}
}
