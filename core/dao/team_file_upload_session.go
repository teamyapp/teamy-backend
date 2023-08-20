package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamFileUploadSession interface {
	FindTeamFileUploadSessionByTeamIDWithTx(
		ct context.Context,
		tx *transaction.Transaction,
		teamID uint64,
		teamFileUploadSessionType entity.TeamFileUploadSessionType,
		fileUploadSessionID uint64,
	) (entity.TeamFileUploadSession, *errs.Error)
	CreateTeamFileUploadSession(ct context.Context, tx *transaction.Transaction, teamFileUploadSession entity.TeamFileUploadSession) *errs.Error
	UpdateTeamFileUploadSession(ct context.Context, tx *transaction.Transaction, teamFileUploadSession entity.TeamFileUploadSession) *errs.Error
}
