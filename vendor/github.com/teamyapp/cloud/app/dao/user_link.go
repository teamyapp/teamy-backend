package dao

import (
	"github.com/teamyapp/cloud/app/entity"
)

type UserLink interface {
	FindByExternalUserID(authProvider string, externalUserID string) (entity.UserLink, error)
	Add(userLink entity.UserLink) error
}
