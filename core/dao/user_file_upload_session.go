package dao

import (
	"github.com/teamyapp/teamy-backend/core/entity"
)

type UserFileUploadSession interface {
	FindUserFileUploadSessionByUserID(
		userID uint64,
		userFileUploadSessionType entity.UserFileUploadSessionType,
		fileUploadSessionID uint64,
	) (entity.UserFileUploadSession, error)
	CreateUserFileUploadSession(userFileUploadSession entity.UserFileUploadSession) error
	UpdateUserFileUploadSession(userFileUploadSession entity.UserFileUploadSession) error
}
