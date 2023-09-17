package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type VersionSelectorVersionRelation interface {
	FindVersionNumbersBySelectorIDWithTx(ct context.Context, tx *transaction.Transaction, versionSelectorID uint64) ([]int, *errs.Error)
	FindVersionNumbersBySelectorID(ct context.Context, selectorID uint64) ([]int, *errs.Error)
	CreateVersionSelectorVersionRelation(ct context.Context, tx *transaction.Transaction, versionSelectorVersionRelation entity.VersionSelectorVersionRelation) *errs.Error
}
