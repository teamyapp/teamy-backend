package dao

import (
	"context"
)

type Thread interface {
	CreateThread(ct context.Context, threadID uint64) error
	DeleteThread(ct context.Context, threadID uint64) error
}
