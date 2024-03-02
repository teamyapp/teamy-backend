package entity

import "time"

type Project struct {
	ID              uint64
	CreatorID       uint64
	Name            string
	ExpectedStartAt *time.Time
	ActualStartAt   *time.Time
	ExpectedEndAt   *time.Time
	ActualEndAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       *time.Time
}
