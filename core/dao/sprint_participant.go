package dao

import (
	"context"

	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintParticipant interface {
	FindParticipantsBySprintID(ct context.Context, sprintID uint64) ([]entity.SprintParticipant, error)
	FindParticipant(ct context.Context, sprintID uint64, participantUserID uint64) (entity.SprintParticipant, error)
	CreateSprintParticipant(ct context.Context, participant entity.SprintParticipant) error
	UpdateSprintParticipant(ct context.Context, participant entity.SprintParticipant) error
	DeleteSprintParticipant(ct context.Context, sprintID uint64, userID uint64) error
}
