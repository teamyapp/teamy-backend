package entity

import (
	"time"
)

type GithubAppInstallation struct {
	ID        string
	TeamID    uint64
	CreatedAt time.Time
}
