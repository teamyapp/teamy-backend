package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppVersionPrice struct{}

// FindAppVersionPricesByAppIDAndVersionNumber implements dao.AppVersionPrice.
func (*AppVersionPrice) FindAppVersionPricesByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int) ([]entity.Money, *errs.Error) {
	panic("unimplemented")
}

var _ dao.AppVersionPrice = (*AppVersionPrice)(nil)

func NewAppVersionPrice() *AppVersionPrice {
	return &AppVersionPrice{}
}
