package entity

import (
	"time"

	oneEntity "github.com/teamyapp/one/entity"
)

type Invitation struct {
	ID             oneEntity.ID
	SenderUserID   oneEntity.ID
	ReceiverUserID *oneEntity.ID
	TeamID         oneEntity.ID
	ExpireAt       time.Time
	Status         InvitationStatus
	Code           string
	CreatedAt      time.Time
}
