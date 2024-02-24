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

type UpdateAppVersion struct {
	logger            telemetry.Logger
	stateSyncer       *realtime.StateSyncer
	appVersionDao     dao.AppVersion
	appDao            dao.App
	id                uint64
	appVersion        entity.AppVersion
	clientNotifiers   []*realtime.ClientNotifier
	notifiersPrepared bool
}

var _ realtime.Mutation = (*UpdateAppVersion)(nil)

func (u *UpdateAppVersion) GetID() uint64 {
	return u.id
}

func (u *UpdateAppVersion) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return u.appVersionDao.UpdateAppVersion(ct, tx, u.appVersion)
}

func (u *UpdateAppVersion) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if u.notifiersPrepared {
		return nil
	}

	var err *errs.Error
	app, err := u.appDao.FindAppByIDWithTx(ct, tx, u.appVersion.AppID)
	if err != nil {
		return err
	}

	u.clientNotifiers, err = u.stateSyncer.GetClientNotifiersByTeamID(ct, app.ManagedByTeamID)
	if err != nil {
		return err
	}

	u.notifiersPrepared = true
	return nil
}

func (u *UpdateAppVersion) Undo() *errs.Error {
	return nil
}

func (u *UpdateAppVersion) GetClientNotifiers() []*realtime.ClientNotifier {
	return u.clientNotifiers
}

func (u *UpdateAppVersion) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.AppVersionCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.appVersion,
	}
}

func (u *UpdateAppVersion) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewUpdateAppVersion(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	appVersionDao dao.AppVersion,
	appDao dao.App,
	appVersion entity.AppVersion,
) *UpdateAppVersion {
	return &UpdateAppVersion{
		logger:        logger,
		stateSyncer:   stateSyncer,
		appVersionDao: appVersionDao,
		appDao:        appDao,
		id:            stateSyncer.NextMutationID(),
		appVersion:    appVersion,
	}
}
