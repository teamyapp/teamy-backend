package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintParticipant interface {
	FindParticipantIDsBySprintID(ct context.Context, sprintID uint64) ([]uint64, *errs.Error)
	FindParticipantsBySprintID(ct context.Context, sprintID uint64) ([]entity.SprintParticipant, *errs.Error)
	FindParticipant(ct context.Context, sprintID uint64, participantUserID uint64) (entity.SprintParticipant, *errs.Error)
	CreateSprintParticipant(ct context.Context, participant entity.SprintParticipant) *errs.Error
	UpdateSprintParticipant(ct context.Context, participant entity.SprintParticipant) *errs.Error
	DeleteSprintParticipant(ct context.Context, sprintID uint64, userID uint64) *errs.Error
}
