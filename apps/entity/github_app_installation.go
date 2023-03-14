package entity

import (
	"time"
)

type GithubAppInstallation struct {
	ID        uint64
	TeamID    uint64
	CreatedAt time.Time
}
