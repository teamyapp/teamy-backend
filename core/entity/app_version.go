package entity

import "time"

type AppVersion struct {
	AppID                     uint64
	VersionNumber             int32
	IconURL                   *string
	HasUIExtension            bool
	UIExtensionEntrypointPath *string
	IsPublic                  bool
	Changes                   *string
	CreatedAt                 time.Time
	UpdateAt                  *time.Time
}
