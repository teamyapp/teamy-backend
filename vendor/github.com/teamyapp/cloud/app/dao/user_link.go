package dao

import (
	"github.com/teamyapp/cloud/app/entity"
)

type UserLink interface {
	FindUserLinkByExternalUserID(authProvider string, externalUserID string) (entity.UserLink, error)
	FindUserLinksByInternalUserID(internalUserID uint64) ([]entity.UserLink, error)
	CreateUserLink(userLink entity.UserLink) error
	DeleteUserLink(authProvider string, internalUserID uint64) error
}
