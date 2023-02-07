package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type UserFileUploadSession interface {
	FindUserFileUploadSessionByUserID(
		ct context.Context,
		userID uint64,
		userFileUploadSessionType entity.UserFileUploadSessionType,
		fileUploadSessionID uint64,
	) (entity.UserFileUploadSession, *errs.Error)
	CreateUserFileUploadSession(ct context.Context, userFileUploadSession entity.UserFileUploadSession) *errs.Error
	UpdateUserFileUploadSession(ct context.Context, userFileUploadSession entity.UserFileUploadSession) *errs.Error
}
