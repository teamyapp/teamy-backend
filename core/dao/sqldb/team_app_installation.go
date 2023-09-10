package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamAppInstallation struct{}

// CreateTeamAppInstallation implements dao.TeamAppInstallation.
func (*TeamAppInstallation) CreateTeamAppInstallation(ct context.Context, teamAppInstallation entity.TeamAppInstallation) (entity.TeamAppInstallation, *errs.Error) {
	panic("unimplemented")
}

// DeleteTeamAppInstallationByIDWithTx implements dao.TeamAppInstallation.
func (*TeamAppInstallation) DeleteTeamAppInstallationByIDWithTx(ct context.Context, tx *transaction.Transaction, appInstallationID uint64) *errs.Error {
	panic("unimplemented")
}

// FindTeamAppInstallationByIDWithTx implements dao.TeamAppInstallation.
func (*TeamAppInstallation) FindTeamAppInstallationByIDWithTx(ct context.Context, tx *transaction.Transaction, appInstallationID uint64) (entity.TeamAppInstallation, *errs.Error) {
	panic("unimplemented")
}

// FindTeamAppInstallationsByAppID implements dao.TeamAppInstallation.
func (*TeamAppInstallation) FindTeamAppInstallationsByAppID(ct context.Context, appID uint64) ([]entity.TeamAppInstallation, *errs.Error) {
	panic("unimplemented")
}

var _ dao.TeamAppInstallation = (*TeamAppInstallation)(nil)

func NewTeamAppInstallation() *TeamAppInstallation {
	return &TeamAppInstallation{}
}
