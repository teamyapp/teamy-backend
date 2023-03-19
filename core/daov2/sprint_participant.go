package daov2

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintParticipant interface {
	FindParticipantsBySprintID(ct context.Context, sprintID uint64) ([]entity.SprintParticipant, *errs.Error)
	FindParticipantIDsBySprintIDWithTx(ct context.Context, tx *transaction.Transaction, sprintID uint64) ([]uint64, *errs.Error)
	FindParticipantsBySprintIDWithTx(ct context.Context, tx *transaction.Transaction, sprintID uint64) ([]entity.SprintParticipant, *errs.Error)
	FindParticipantWithTx(ct context.Context, tx *transaction.Transaction, sprintID uint64, participantUserID uint64) (entity.SprintParticipant, *errs.Error)
	CreateSprintParticipant(ct context.Context, tx *transaction.Transaction, participant entity.SprintParticipant) *errs.Error
	UpdateSprintParticipant(ct context.Context, tx *transaction.Transaction, participant entity.SprintParticipant) *errs.Error
	DeleteSprintParticipant(ct context.Context, tx *transaction.Transaction, sprintID uint64, userID uint64) *errs.Error
}
