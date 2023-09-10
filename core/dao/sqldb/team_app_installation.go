package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamAppInstallation struct{}

var _ dao.TeamAppInstallation = (*TeamAppInstallation)(nil)

func (*TeamAppInstallation) FindTeamAppInstallationByIDWithTx(ct context.Context, tx *transaction.Transaction, appInstallationID uint64) (entity.TeamAppInstallation, *errs.Error) {
	panic("unimplemented")
}

func (*TeamAppInstallation) FindTeamAppInstallationsByAppID(ct context.Context, appID uint64) ([]entity.TeamAppInstallation, *errs.Error) {
	panic("unimplemented")
}
func (*TeamAppInstallation) CreateTeamAppInstallation(ct context.Context, teamAppInstallation entity.TeamAppInstallation) (entity.TeamAppInstallation, *errs.Error) {
	panic("unimplemented")
}

func (*TeamAppInstallation) DeleteTeamAppInstallationByIDWithTx(ct context.Context, tx *transaction.Transaction, appInstallationID uint64) *errs.Error {
	panic("unimplemented")
}

func NewTeamAppInstallation() *TeamAppInstallation {
	return &TeamAppInstallation{}
}
