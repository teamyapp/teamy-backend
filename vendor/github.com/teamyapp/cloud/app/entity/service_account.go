package entity

import (
	"time"
)

type ServiceAccount struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Secret      *string   `json:"secret"`
	OwnerUserID uint64    `json:"owner_user_id"`
	CreatedAt   time.Time `json:"created_at"`
}
