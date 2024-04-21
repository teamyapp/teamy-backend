package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type VersionSelector interface {
	FindVersionSelectorByIDWithTx(ct context.Context, tx *transaction.Transaction, versionSelectorID uint64) (entity.VersionSelector, *errs.Error)
	FindVersionSelectorByID(ct context.Context, selectorID uint64) (entity.VersionSelector, *errs.Error)
	CreateVersionSelector(ct context.Context, tx *transaction.Transaction, versionSelector entity.VersionSelector) *errs.Error
	UpdateVersionSelector(ct context.Context, tx *transaction.Transaction, versionSelector entity.VersionSelector) *errs.Error
	DeleteVersionSelector(ct context.Context, tx *transaction.Transaction, versionSelectorID uint64) *errs.Error
}
