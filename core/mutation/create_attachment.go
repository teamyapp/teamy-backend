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

type CreateAttachment struct {
	logger            telemetry.Logger
	stateSyncer       *realtime.StateSyncer
	attachmentListDao dao.AttachmentList
	attachmentDao     dao.Attachment
	taskDao           dao.Task
	id                uint64
	attachment        entity.Attachment
	clientNotifiers   []*realtime.ClientNotifier
	notifiersPrepared bool
}

var _ realtime.Mutation = (*CreateAttachment)(nil)

func (c *CreateAttachment) GetID() uint64 {
	return c.id
}

func (c *CreateAttachment) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	internalErr := c.attachmentDao.CreateAttachment(ct, tx, c.attachment)
	return internalErr
}

func (c *CreateAttachment) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifiersPrepared {
		return nil
	}

	attachmentList, internalErr := c.attachmentListDao.FindAttachmentListByIDWithTx(ct, tx, c.attachment.AttachmentListID)
	if internalErr != nil {
		return internalErr
	}

	switch attachmentList.OwnerType {
	case entity.AttachmentListOwnerTypeTask:
		task, internalErr := c.taskDao.FindTaskByIDWithTx(ct, tx, attachmentList.OwnerID)
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

func (c *CreateAttachment) Undo() *errs.Error {
	return nil
}

func (c *CreateAttachment) GetClientNotifiers() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateAttachment) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.ImageCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.attachment,
	}
}

func (c *CreateAttachment) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateAttachment(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	attachmentListDao dao.AttachmentList,
	attachmentDao dao.Attachment,
	taskDao dao.Task,
	attachment entity.Attachment,
) *CreateAttachment {
	return &CreateAttachment{
		logger:            logger,
		stateSyncer:       stateSyncer,
		attachmentListDao: attachmentListDao,
		attachmentDao:     attachmentDao,
		taskDao:           taskDao,
		attachment:        attachment,
	}
}
