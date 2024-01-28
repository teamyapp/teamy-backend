package entity

import (
	"time"
)

type AppSecret struct {
	ID            uint64
	Name          string
	Token         string
	Secret        string
	AddedAt       time.Time
	AddedByUserID uint64
	LastUsedAt    *time.Time
	AppID         uint64
}
