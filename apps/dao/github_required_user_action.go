package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubRequiredUserAction interface {
	FindRequiredUserActionsByActionUserID(
		ct context.Context,
		teamID uint64,
		actionUserID uint64,
	) ([]entity.GithubRequiredUserAction, *errs.Error)
	FindRequiredUserActionByActionTypeAndUserID(
		ct context.Context,
		teamID uint64,
		actionType entity.GithubUserActionType,
		actionUserID uint64,
	) (entity.GithubRequiredUserAction, *errs.Error)
	CreateRequiredUserAction(ct context.Context, requiredUserAction entity.GithubRequiredUserAction) *errs.Error
	UpdateRequiredUserAction(ct context.Context, requiredUserAction entity.GithubRequiredUserAction) *errs.Error
	DeleteRequiredUserActionsByActionTypeAndUserID(
		ct context.Context,
		actionType entity.GithubUserActionType,
		actionUserID uint64,
	) *errs.Error
}
