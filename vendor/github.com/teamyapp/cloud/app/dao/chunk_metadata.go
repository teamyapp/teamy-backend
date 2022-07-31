package dao

import (
	"github.com/teamyapp/cloud/app/entity"
)

type ChunkMetadata interface {
	FindChunkMetadataID(chunkID uint64) (entity.ChunkMetadata, error)
	CreateChunkMetadata(metadata entity.ChunkMetadata) error
	UpdateChunkMetadata(metadata entity.ChunkMetadata) error
}
