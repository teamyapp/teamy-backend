package entity

import "time"

type AppVersion struct {
	AppID                     uint64
	VersionNumber             int32
	IconUrl                   *string
	HasUiExtension            bool
	UiExtensionEntrypointPath *string
	IsPublic                  bool
	Changes                   *string
	CreatedAt                 time.Time
	UpdateAt                  *time.Time
}
