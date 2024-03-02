package entity

import "time"

type PhaseStatus string

const (
	TodoPhaseStatus       PhaseStatus = "TODO"
	InProgressPhaseStatus PhaseStatus = "IN_PROGRESS"
	PausedPhaseStatus     PhaseStatus = "PAUSED"
	CompletedPhaseStatus  PhaseStatus = "COMPLETED"
)

type Phase struct {
	ID              uint64
	Name            string
	Status          PhaseStatus
	ExpectedStartAt time.Time
	ActualStartAt   *time.Time
	ExpectedEndAt   time.Time
	ActualEndAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       *time.Time
}
