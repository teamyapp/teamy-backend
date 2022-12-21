package collection

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type InvitationSyncer struct {
	dataCollector       obs.DataCollector
	realTimeStateSyncer *realtime.StateSyncer
	invitationDao       dao.Invitation
}

func (i InvitationSyncer) CreateAndSyncInvitation(ct context.Context, tx realtime.Transaction, invitation entity.Invitation) error {
	err := i.invitationDao.CreateInvitation(ct, invitation)
	if err != nil {
		i.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx.AddMutation(ct, realtime.MutationInput{
		CollectionType: realtime.InvitationCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        invitation,
	})
	return nil
}

func (i InvitationSyncer) UpdateAndSyncInvitation(ct context.Context, tx realtime.Transaction, invitation entity.Invitation) error {
	err := i.invitationDao.UpdateInvitation(ct, invitation)
	if err != nil {
		i.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx.AddMutation(ct, realtime.MutationInput{
		CollectionType: realtime.InvitationCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        invitation,
	})
	return nil
}

func (i InvitationSyncer) DeleteAndSyncInvitation(ct context.Context, tx realtime.Transaction, invitationID uint64) error {

	err := i.invitationDao.DeleteInvitation(ct, invitationID)
	if err != nil {
		i.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx.AddMutation(ct, realtime.MutationInput{
		CollectionType: realtime.InvitationCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload:        invitationID,
	})
	return nil
}

func NewInvitationSyncer(dataCollector obs.DataCollector, realTimeStateSyncer *realtime.StateSyncer, invitationDao dao.Invitation) InvitationSyncer {
	return InvitationSyncer{
		dataCollector:       dataCollector,
		realTimeStateSyncer: realTimeStateSyncer,
		invitationDao:       invitationDao,
	}
}
