package collection

import (
	"strconv"

	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/infras/storage"
)

const invitationCollectionType = "Invitation"

type InvitationSyncer struct {
	realTimeCollection *storage.RealTimeCollections
	invitationDao      dao.Invitation
}

func (i InvitationSyncer) CreateAndSyncInvitation(invitation entity.Invitation) error {
	err := i.invitationDao.CreateInvitation(invitation)
	if err != nil {
		return err
	}

	return i.realTimeCollection.Mutate(storage.Mutation{
		CollectionType: invitationCollectionType,
		MutationType:   storage.CreateMutationType,
		Attributes:     storage.MapAttributes(invitation),
	})
}

func (i InvitationSyncer) UpdateAndSyncInvitation(invitation entity.Invitation) error {
	err := i.invitationDao.UpdateInvitation(invitation)
	if err != nil {
		return err
	}

	return i.realTimeCollection.Mutate(storage.Mutation{
		CollectionType: invitationCollectionType,
		MutationType:   storage.UpdateMutationType,
		Attributes:     storage.MapAttributes(invitation),
	})
}

func (i InvitationSyncer) DeleteAndSyncInvitation(invitationID uint64) error {
	err := i.invitationDao.DeleteInvitation(invitationID)
	if err != nil {
		return err
	}

	idStr := strconv.FormatUint(invitationID, 10)
	return i.realTimeCollection.Mutate(storage.Mutation{
		CollectionType: invitationCollectionType,
		MutationType:   storage.DeleteMutationType,
		Attributes: map[string]*string{
			"ID": &idStr,
		},
	})
}

func NewInvitationSyncer(realTimeCollection *storage.RealTimeCollections, invitationDao dao.Invitation) InvitationSyncer {
	realTimeCollection.RegisterCollectionType(invitationCollectionType)
	return InvitationSyncer{
		realTimeCollection: realTimeCollection,
		invitationDao:      invitationDao,
	}
}
