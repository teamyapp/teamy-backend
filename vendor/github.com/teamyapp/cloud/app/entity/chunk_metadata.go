package entity

import (
	"time"
)

type ChunkMetadata struct {
	ID          uint64    `json:"id"`
	SizeInBytes uint64    `json:"sizeInBytes"`
	CreatedAt   time.Time `json:"createdAt"`
}
