package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppVersionPrice interface {
	FindAppVersionPricesByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int) ([]entity.Money, *errs.Error)
}
