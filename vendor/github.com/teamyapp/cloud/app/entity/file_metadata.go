package entity

import (
	"time"
)

type FileMetadata struct {
	ID             uint64     `json:"id"`
	Name           string     `json:"name"`
	SizeInBytes    uint64     `json:"sizeInBytes"`
	MIMEType       string     `json:"mimeType"`
	ChunkIDs       []uint64   `json:"chunkIds"`
	CreatedAt      time.Time  `json:"createdAt"`
	LastModifiedAt *time.Time `json:"lastModifiedAt"`
}
