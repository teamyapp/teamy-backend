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
	AppVersionStatusError      AppVersionStatus = "ERROR"
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
	ErrorMessage    *string
}

type AppVersionChange struct {
	ID            uint64
	AppID         uint64
	VersionNumber int
	Change        string
}

type AppVersionPrice struct {
	Money
	AppID         uint64
	VersionNumber int
}
