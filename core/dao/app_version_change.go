package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
)

type AppVersionChange interface {
	FindAppVersionChangesByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int) ([]string, *errs.Error)
}
