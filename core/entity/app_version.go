package entity

import (
	"time"
)

type AppVersion struct {
	AppID           uint64
	Number          int
	AppName         string
	Description     string
	HasUiExtension  bool
	CreatedAt       time.Time
	CreatedByUserID uint64
	IsReady         bool
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
