package entity

import (
	"time"
)

type GithubAppInstallation struct {
	ID        int
	TeamID    uint64
	CreatedAt time.Time
}
