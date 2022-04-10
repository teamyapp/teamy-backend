package dao

import "github.com/teamyapp/teamy-backend/app/entityv2"

type Invitation interface {
	FindInvitation(id uint64) (entityv2.Invitation, error)
}
