package entity

import (
	"time"

	oneEntity "github.com/teamyapp/one/entity"
)

type Invitation struct {
	oneEntity.Entity
	InviterUserID   oneEntity.ID
	NewMemberUserID oneEntity.ID
	NewMemberEmail  string
	TeamID          oneEntity.ID
	TTL             time.Duration
	Status          InvitationStatus
}
