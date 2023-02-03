package entity

import "time"

type AppTeamInstallation struct {
	AppID                uint64
	InstalledTeamID      uint64
	InstalledByUserID    *uint64
	EnabledVersionNumber int32
	InstalledAt          time.Time
}
