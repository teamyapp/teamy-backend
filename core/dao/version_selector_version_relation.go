package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type VersionSelectorVersionRelation interface {
	FindVersionNumbersBySelectorID(ct context.Context, selectorID uint64) ([]int, *errs.Error)
	CreateVersionSelectorVersionRelation(ct context.Context, versionSelectorVersionRelation entity.VersionSelectorVersionRelation) (VersionSelectorVersionRelation, *errs.Error)
}
