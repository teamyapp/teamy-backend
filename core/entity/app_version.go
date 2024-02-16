package entity

import (
	"time"
)

type AppVersionStatus string

const (
	AppVersionStatusInit       AppVersionStatus = "INIT"
	AppVersionStatusUploading  AppVersionStatus = "UPLOADING"
	AppVersionStatusProcessing AppVersionStatus = "PROCESSING"
	AppVersionStatusReady      AppVersionStatus = "READY"
)

type AppVersion struct {
	AppID           uint64
	Number          int
	AppName         string
	Description     string
	HasUiExtension  bool
	CreatedAt       time.Time
	UpdatedAt       *time.Time
	CreatedByUserID uint64
	Status          AppVersionStatus
	Locked          bool
	IconURL         *string
}

type AppVersionChange struct {
	AppID         uint64
	ChangeID      uint64
	VersionNumber int
	Change        string
}

type AppVersionPrice struct {
	Money
	AppID         uint64
	VersionNumber int
}
