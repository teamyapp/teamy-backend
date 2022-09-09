package dao

import (
	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintParticipant interface {
	FindParticipantsBySprintID(sprintID uint64) ([]entity.SprintParticipant, error)
	FindParticipant(sprintID uint64, participantUserID uint64) (entity.SprintParticipant, error)
	CreateSprintParticipant(participant entity.SprintParticipant) error
	UpdateSprintParticipant(participant entity.SprintParticipant) error
	DeleteSprintParticipant(sprintID uint64, userID uint64) error
}
