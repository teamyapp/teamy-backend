package entity

import (
	"time"
)

type GithubAppInstallState struct {
	ID          uint64
	TeamID      uint64
	RedirectURL string
	CreatedAt   time.Time
}
