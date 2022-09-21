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

func (i InvitationSyncer) CreateAndSyncInvitation(ct context.Context, invitation entity.Invitation) error {
	err := i.invitationDao.CreateInvitation(ct, invitation)
	if err != nil {
		i.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	i.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.InvitationCollectionType,
		MutationType:   realtime.CreateMutationType,
		TeamIDs: []uint64{
			invitation.TeamID,
		},
		Payload: invitation,
	})
	return nil
}

func (i InvitationSyncer) UpdateAndSyncInvitation(ct context.Context, invitation entity.Invitation) error {
	err := i.invitationDao.UpdateInvitation(ct, invitation)
	if err != nil {
		i.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	i.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.InvitationCollectionType,
		MutationType:   realtime.UpdateMutationType,
		TeamIDs: []uint64{
			invitation.TeamID,
		},
		Payload: invitation,
	})
	return nil
}

func (i InvitationSyncer) DeleteAndSyncInvitation(ct context.Context, invitationID uint64) error {
	invitation, err := i.invitationDao.FindInvitationByID(ct, invitationID)
	if err != nil {
		i.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	err = i.invitationDao.DeleteInvitation(ct, invitationID)
	if err != nil {
		i.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	i.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.InvitationCollectionType,
		MutationType:   realtime.DeleteMutationType,
		TeamIDs: []uint64{
			invitation.TeamID,
		},
		Payload: invitationID,
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
