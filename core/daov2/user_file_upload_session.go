package daov2

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type UserFileUploadSession interface {
	FindUserFileUploadSessionByUserIDWithTx(
		ct context.Context,
		tx *transaction.Transaction,
		userID uint64,
		userFileUploadSessionType entity.UserFileUploadSessionType,
		fileUploadSessionID uint64,
	) (entity.UserFileUploadSession, *errs.Error)
	CreateUserFileUploadSession(ct context.Context, tx *transaction.Transaction, userFileUploadSession entity.UserFileUploadSession) *errs.Error
	UpdateUserFileUploadSession(ct context.Context, tx *transaction.Transaction, userFileUploadSession entity.UserFileUploadSession) *errs.Error
}
