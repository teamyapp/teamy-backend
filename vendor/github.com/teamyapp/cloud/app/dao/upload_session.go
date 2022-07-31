package dao

import (
	"github.com/teamyapp/cloud/app/entity"
)

type UploadSession interface {
	FindUploadSessionByID(uploadSessionID uint64) (entity.UploadSession, error)
	CreateUploadSession(uploadSession entity.UploadSession) error
	UpdateUploadSession(uploadSession entity.UploadSession) error
}
