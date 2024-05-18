package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateAttachmentList struct {
	logger            telemetry.Logger
	stateSyncer       *realtime.StateSyncer
	attachmentListDao dao.AttachmentList
	taskDao           dao.Task
	id                uint64
	attachmentList    entity.AttachmentList
	clientNotifiers   []*realtime.ClientNotifier
	notifiersPrepared bool
}

var _ realtime.Mutation = (*CreateAttachmentList)(nil)

func (c *CreateAttachmentList) GetID() uint64 {
	return c.id
}

func (c *CreateAttachmentList) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	internalErr := c.attachmentListDao.CreateAttachmentList(ct, tx, c.attachmentList)
	return internalErr
}

func (c *CreateAttachmentList) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifiersPrepared {
		return nil
	}

	switch c.attachmentList.OwnerType {
	case entity.AttachmentListOwnerTypeTask:
		task, internalErr := c.taskDao.FindTaskByIDWithTx(ct, tx, c.attachmentList.OwnerID)
		if internalErr != nil {
			return internalErr
		}

		c.clientNotifiers, internalErr = c.stateSyncer.GetClientNotifiersByTeamID(ct, task.OwningTeamID)
		if internalErr != nil {
			return internalErr
		}

		c.notifiersPrepared = true
	}

	return nil
}

func (c *CreateAttachmentList) Undo() *errs.Error {
	return nil
}

func (c *CreateAttachmentList) GetClientNotifiers() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateAttachmentList) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.AttachmentListCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.attachmentList,
	}
}

func (c *CreateAttachmentList) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateAttachmentList(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	attachmentListDao dao.AttachmentList,
	taskDao dao.Task,
	attachmentList entity.AttachmentList,
) *CreateAttachmentList {
	return &CreateAttachmentList{
		logger:            logger,
		stateSyncer:       stateSyncer,
		attachmentListDao: attachmentListDao,
		taskDao:           taskDao,
		attachmentList:    attachmentList,
	}
}
