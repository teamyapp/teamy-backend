package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type VersionSelector struct{}

// CreateVersionSelector implements dao.VersionSelector.
func (*VersionSelector) CreateVersionSelector(ct context.Context, versionSelector entity.VersionSelector) (entity.VersionSelector, *errs.Error) {
	panic("unimplemented")
}

// FindVersionSelectorByID implements dao.VersionSelector.
func (*VersionSelector) FindVersionSelectorByID(ct context.Context, selectorID uint64) (entity.VersionSelector, *errs.Error) {
	panic("unimplemented")
}

var _ dao.VersionSelector = (*VersionSelector)(nil)

func NewVersionSelector() *VersionSelector {
	return &VersionSelector{}
}
