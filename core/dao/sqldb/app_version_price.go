package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppVersionPrice struct{}

var _ dao.AppVersionPrice = (*AppVersionPrice)(nil)

func (*AppVersionPrice) FindAppVersionPricesByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int) ([]entity.Money, *errs.Error) {
	panic("unimplemented")
}

func NewAppVersionPrice() *AppVersionPrice {
	return &AppVersionPrice{}
}
