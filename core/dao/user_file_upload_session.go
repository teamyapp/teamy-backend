package dao

import (
	"context"

	"github.com/teamyapp/teamy-backend/core/entity"
)

type UserFileUploadSession interface {
	FindUserFileUploadSessionByUserID(
		ct context.Context,
		userID uint64,
		userFileUploadSessionType entity.UserFileUploadSessionType,
		fileUploadSessionID uint64,
	) (entity.UserFileUploadSession, error)
	CreateUserFileUploadSession(ct context.Context, userFileUploadSession entity.UserFileUploadSession) error
	UpdateUserFileUploadSession(ct context.Context, userFileUploadSession entity.UserFileUploadSession) error
}
