package datastore

import (
	"fmt"
	"time"

	"github.com/teamyapp/teamy-backend/app/entity"
)

func (d DataStore) CreateInvitation(invitation entity.Invitation) (entity.Invitation, error) {
	invitation.ID = d.newID(Invitation)
	invitation.CreatedAt = time.Now()
	d.data.Invitations[invitation.ID] = invitation
	return invitation, d.persister.Write(d.data)
}

func (d DataStore) FilterInvitations(filter func(invitation entity.Invitation) bool) []entity.Invitation {
	invitations := make([]entity.Invitation, 0)
	for _, invitation := range d.data.Invitations {
		if filter(invitation) {
			invitations = append(invitations, invitation)
		}
	}
	return invitations
}

func (d DataStore) UpdateInvitation(
	invitationID uint64,
	apply func(invitation entity.Invitation) entity.Invitation,
) (entity.Invitation, error) {
	val, ok := d.data.Invitations[invitationID]
	if !ok {
		return entity.Invitation{}, fmt.Errorf("invitation not found: %v", invitationID)
	}

	invitation := apply(val)
	d.data.Invitations[invitationID] = invitation
	return invitation, d.persister.Write(d.data)
}

func (d DataStore) DeleteInvitation(id uint64) (entity.Invitation, error) {
	val, ok := d.data.Invitations[id]
	if !ok {
		return entity.Invitation{}, fmt.Errorf("invitation not found: %v", id)
	}

	delete(d.data.Invitations, id)
	return val, d.persister.Write(d.data)
}
