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

type CreateImage struct {
	logger            telemetry.Logger
	stateSyncer       *realtime.StateSyncer
	attachmentListDao dao.AttachmentList
	imageDao          dao.Image
	taskDao           dao.Task
	id                uint64
	image             entity.Image
	clientNotifiers   []*realtime.ClientNotifier
	notifiersPrepared bool
}

var _ realtime.Mutation = (*CreateImage)(nil)

func (c *CreateImage) GetID() uint64 {
	return c.id
}

func (c *CreateImage) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	internalErr := c.imageDao.CreateImage(ct, tx, c.image)
	return internalErr
}

func (c *CreateImage) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifiersPrepared {
		return nil
	}

	attachmentList, internalErr := c.attachmentListDao.FindAttachmentListByIDWithTx(ct, tx, c.image.AttachmentListID)
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

func (c *CreateImage) Undo() *errs.Error {
	return nil
}

func (c *CreateImage) GetClientNotifiers() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateImage) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.ImageCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.image,
	}
}

func (c *CreateImage) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateImage(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	attachmentListDao dao.AttachmentList,
	imageDao dao.Image,
	taskDao dao.Task,
	image entity.Image,
) *CreateImage {
	return &CreateImage{
		logger:            logger,
		stateSyncer:       stateSyncer,
		attachmentListDao: attachmentListDao,
		imageDao:          imageDao,
		taskDao:           taskDao,
		image:             image,
	}
}
