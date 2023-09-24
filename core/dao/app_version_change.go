package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
)

type AppVersionChange interface {
	FindAppVersionChangesByAppIDAndVersionNumberWithTx(ct context.Context, tx *transaction.Transaction, appID uint64, versionNumber int) ([]string, *errs.Error)
	FindAppVersionChangesByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int) ([]string, *errs.Error)
}
