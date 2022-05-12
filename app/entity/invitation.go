package entity

import (
	"time"
)

type Invitation struct {
	ID                uint64
	SenderUserID      uint64
	ReceiverFirstName string
	ReceiverLastName  string
	ReceiverUserID    *uint64
	TeamID            uint64
	ExpireAt          time.Time
	Status            InvitationStatus
	Code              string
	CreatedAt         time.Time
}
