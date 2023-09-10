package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type VersionSelector interface {
	CreateVersionSelector(ct context.Context, versionSelector entity.VersionSelector) (entity.VersionSelector, *errs.Error)
	FindVersionSelectorByID(ct context.Context, selectorID uint64) (entity.VersionSelector, *errs.Error)
}
