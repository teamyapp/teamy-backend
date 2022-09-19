package collection

import (
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

func (i InvitationSyncer) CreateAndSyncInvitation(invitation entity.Invitation) error {
	err := i.invitationDao.CreateInvitation(invitation)
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	i.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.InvitationCollectionType,
		MutationType:   realtime.CreateMutationType,
		TeamIDs: []uint64{
			invitation.TeamID,
		},
		Payload: invitation,
	},
	)
	return nil
}

func (i InvitationSyncer) UpdateAndSyncInvitation(invitation entity.Invitation) error {
	err := i.invitationDao.UpdateInvitation(invitation)
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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

func (i InvitationSyncer) DeleteAndSyncInvitation(invitationID uint64) error {
	invitation, err := i.invitationDao.FindInvitationByID(invitationID)
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	err = i.invitationDao.DeleteInvitation(invitationID)
	if err != nil {
		i.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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
