package entity

import (
	"time"
)

type AppVersion struct {
	AppID           uint64
	Number          int
	AppName         string
	Description     string
	CreatedAt       time.Time
	CreatedByUserID uint64
	IsReady         bool
}

type AppVersionChange struct {
	AppID         uint64
	VersionNumber int
	Change        string
}

type AppVersionPrice struct {
	Money
	appID         uint64
	VersionNumber int
}
