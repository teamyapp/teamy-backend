package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type VersionSelectorVersionRelation interface {
	FindVersionNumbersBySelectorID(ct context.Context, selectorID uint64) ([]int, *errs.Error)
	CreateVersionSelectorVersionRelation(ct context.Context, tx *transaction.Transaction, versionSelectorVersionRelation entity.VersionSelectorVersionRelation) *errs.Error
}
