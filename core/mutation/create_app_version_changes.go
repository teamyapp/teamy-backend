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

type CreateAppVersionChange struct {
	logger              telemetry.Logger
	stateSyncer         *realtime.StateSyncer
	id                  uint64
	clientNotifiers     []*realtime.ClientNotifier
	notifiersPrepared   bool
	appVersionChange    entity.AppVersionChange
	appVersionChangeDao dao.AppVersionChange
	appDao              dao.App
}

var _ realtime.Mutation = (*CreateAppVersionChange)(nil)

func (c *CreateAppVersionChange) GetID() uint64 {
	return c.id
}

func (c *CreateAppVersionChange) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return c.appVersionChangeDao.CreateAppVersionChange(ct, tx, c.appVersionChange)
}

func (c *CreateAppVersionChange) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifiersPrepared {
		return nil
	}

	var err *errs.Error
	app, err := c.appDao.FindAppByIDWithTx(ct, tx, c.appVersionChange.AppID)
	if err != nil {
		return err
	}

	c.clientNotifiers, err = c.stateSyncer.GetClientNotifiersByTeamID(ct, app.ManagedByTeamID)
	if err != nil {
		return err
	}

	c.notifiersPrepared = true
	return nil
}

func (c *CreateAppVersionChange) Undo() *errs.Error {
	return nil
}

func (c *CreateAppVersionChange) GetClientNotifiers() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateAppVersionChange) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.AppVersionChangeCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.appVersionChange,
	}
}

func (c *CreateAppVersionChange) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateAppVersionChange(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	appVersionChange entity.AppVersionChange,
	appVersionChangeDao dao.AppVersionChange,
	appDao dao.App,
) *CreateAppVersionChange {
	return &CreateAppVersionChange{
		logger:              logger,
		stateSyncer:         stateSyncer,
		id:                  stateSyncer.NextMutationID(),
		notifiersPrepared:   false,
		appVersionChange:    appVersionChange,
		appVersionChangeDao: appVersionChangeDao,
		appDao:              appDao,
	}
}
