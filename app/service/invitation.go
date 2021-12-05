package service

import (
	"time"

	oneEntity "github.com/teamyapp/one/entity"
)

type Invitation struct {
	sendInviteUserID    oneEntity.ID
	receiveInviteUserID string
	link                string
	ttl                 time.Duration
	teamID              oneEntity.ID
}
