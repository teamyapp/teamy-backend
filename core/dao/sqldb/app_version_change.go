package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
)

type AppVersionChange struct {
}

// FindAppVersionChangesByAppIDAndVersionNumber implements dao.AppVersionChange.
func (*AppVersionChange) FindAppVersionChangesByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int) ([]string, *errs.Error) {
	panic("unimplemented")
}

var _ dao.AppVersionChange = (*AppVersionChange)(nil)

func NewAppVersionChange() *AppVersionChange {
	return &AppVersionChange{}
}
