package dao

import (
	"context"

	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubRequiredUserAction interface {
	FindRequiredUserActionsByActionUserID(
		ct context.Context,
		teamID uint64,
		actionUserID uint64,
	) ([]entity.GithubRequiredUserAction, error)
	CreateRequiredUserAction(ct context.Context, requiredUserAction entity.GithubRequiredUserAction) error
	UpdateRequiredUserAction(ct context.Context, requiredUserAction entity.GithubRequiredUserAction) error
}
