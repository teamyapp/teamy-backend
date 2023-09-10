package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type VersionSelectorVersionRelation struct{}

// CreateVersionSelectorVersionRelation implements dao.VersionSelectorVersionRelation.
func (*VersionSelectorVersionRelation) CreateVersionSelectorVersionRelation(ct context.Context, versionSelectorVersionRelation entity.VersionSelectorVersionRelation) (dao.VersionSelectorVersionRelation, *errs.Error) {
	panic("unimplemented")
}

// FindVersionNumbersBySelectorID implements dao.VersionSelectorVersionRelation.
func (*VersionSelectorVersionRelation) FindVersionNumbersBySelectorID(ct context.Context, selectorID uint64) ([]int, *errs.Error) {
	panic("unimplemented")
}

var _ dao.VersionSelectorVersionRelation = (*VersionSelectorVersionRelation)(nil)

func NewVersionSelectorVersionRelation() *VersionSelectorVersionRelation {
	return &VersionSelectorVersionRelation{}
}
