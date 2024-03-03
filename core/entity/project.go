package entity

import "time"

type Project struct {
	ID              uint64
	Name            string
	ExpectedStartAt *time.Time
	ActualStartAt   *time.Time
	ExpectedEndAt   *time.Time
	ActualEndAt     *time.Time
	CreatorID       uint64
	CreatedAt       time.Time
	UpdatedAt       *time.Time
}
