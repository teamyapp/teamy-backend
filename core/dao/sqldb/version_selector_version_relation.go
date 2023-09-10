package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type VersionSelectorVersionRelation struct{}

var _ dao.VersionSelectorVersionRelation = (*VersionSelectorVersionRelation)(nil)

func (*VersionSelectorVersionRelation) FindVersionNumbersBySelectorID(ct context.Context, selectorID uint64) ([]int, *errs.Error) {
	panic("unimplemented")
}
func (*VersionSelectorVersionRelation) CreateVersionSelectorVersionRelation(ct context.Context, versionSelectorVersionRelation entity.VersionSelectorVersionRelation) (dao.VersionSelectorVersionRelation, *errs.Error) {
	panic("unimplemented")
}

func NewVersionSelectorVersionRelation() *VersionSelectorVersionRelation {
	return &VersionSelectorVersionRelation{}
}
