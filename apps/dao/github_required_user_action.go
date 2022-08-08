package dao

import (
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubRequiredUserAction interface {
	FindRequiredUserActionsByActionUserID(
		teamID uint64,
		actionUserID uint64,
	) ([]entity.GithubRequiredUserAction, error)
	CreateRequiredUserAction(requiredUserAction entity.GithubRequiredUserAction) error
	UpdateRequiredUserAction(requiredUserAction entity.GithubRequiredUserAction) error
}
