package entity

import (
	"time"
)

type UploadSessionStatus string

const (
	CreatedUploadSessionStatus         UploadSessionStatus = "CREATED"
	InitializedUploadSessionStatus     UploadSessionStatus = "INITIALIZED"
	UploadingChunksUploadSessionStatus UploadSessionStatus = "UPLOADING_CHUNKS"
	CompletedUploadSessionStatus       UploadSessionStatus = "COMPLETED"
)

type UploadSession struct {
	ID                     uint64              `json:"id"`
	Status                 UploadSessionStatus `json:"status"`
	UploadedSizeInBytes    uint64              `json:"uploadedSizeInBytes"`
	FileID                 uint64              `json:"fileId"`
	FileName               string              `json:"fileName"`
	MIMEType               string              `json:"mimeType"`
	TotalSizeInBytes       uint64              `json:"totalSizeInBytes"`
	TotalNumOfChunks       int                 `json:"totalNumOfChunks"`
	ChunkIDs               []uint64            `json:"chunkIds"`
	NextChunkIndexToUpload int                 `json:"nextChunkIndexToUpload"`
	HashState              []byte              `json:"hashState"`
	ActualContentHash      string              `json:"actualContentHash"`
	ExpectedContentHash    string              `json:"expectedContentHash"`
	CreatedAt              time.Time           `json:"createdAt"`
	UpdatedAt              *time.Time          `json:"updatedAt"`
}
