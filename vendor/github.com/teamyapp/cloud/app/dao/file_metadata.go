package dao

import (
	"github.com/teamyapp/cloud/app/entity"
)

type FileMetadata interface {
	FindMetadataByFileID(fileID uint64) (entity.FileMetadata, error)
	CreateFileMetadata(metadata entity.FileMetadata) error
	UpdateFileMetadata(metadata entity.FileMetadata) error
}
