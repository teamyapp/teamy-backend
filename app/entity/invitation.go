package entity

import (
	"time"

	oneEntity "github.com/teamyapp/one/entity"
)

type Invitation struct {
	oneEntity.Entity
	SendInviteUserID       oneEntity.ID
	ReceiveInviteUserID    oneEntity.ID
	ReceiveInviteUserEmail string
	TeamID                 oneEntity.ID
	Expiration             time.Time
	Status                 InvitationStatus
}
