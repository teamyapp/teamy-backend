package mutation

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateSprintParticipantMutation struct {
	dataCollector          telemetry.DataCollector
	stateSyncer            *realtime.StateSyncer
	sprintParticipantDao   dao.SprintParticipant
	sprintParticipantDaoV2 daov2.SprintParticipant
	sprintDao              dao.Sprint
	sprintDaoV2            daov2.Sprint
	id                     uint64
	sprintParticipant      entity.SprintParticipant
	clientNotifiers        []*realtime.ClientNotifier
}

var _ realtime.Mutation = (*UpdateSprintParticipantMutation)(nil)

func (u *UpdateSprintParticipantMutation) GetID() uint64 {
	return u.id
}

func (u *UpdateSprintParticipantMutation) ExecuteV2(ct context.Context, tx *sql.Tx) *errs.Error {
	internalErr := u.sprintParticipantDaoV2.UpdateSprintParticipant(ct, tx, u.sprintParticipant)
	if internalErr != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (u *UpdateSprintParticipantMutation) PrepareClientNotifiers(ct context.Context, tx *sql.Tx) *errs.Error {
	sprint, internalErr := u.sprintDaoV2.FindSprintByID(ct, tx, u.sprintParticipant.SprintID)
	if internalErr != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	u.clientNotifiers, internalErr = u.stateSyncer.GetClientNotifiersByTeamID(ct, sprint.OwningTeamID)
	if internalErr != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (u *UpdateSprintParticipantMutation) Execute(ct context.Context) *errs.Error {
	err := u.sprintParticipantDao.UpdateSprintParticipant(ct, u.sprintParticipant)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (u *UpdateSprintParticipantMutation) Undo() *errs.Error {
	return nil
}

func (u *UpdateSprintParticipantMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	sprint, err := u.sprintDao.FindSprintByID(ct, u.sprintParticipant.SprintID)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return []*realtime.ClientNotifier{}, err
	}

	return u.stateSyncer.GetClientNotifiersByTeamID(ct, sprint.OwningTeamID)
}

func (u *UpdateSprintParticipantMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return u.clientNotifiers
}

func (u *UpdateSprintParticipantMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.SprintParticipantCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.sprintParticipant,
	}
}

func (u *UpdateSprintParticipantMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewUpdateSprintParticipantMutation(
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	sprintParticipantDao dao.SprintParticipant,
	sprintParticipantDaoV2 daov2.SprintParticipant,
	sprintDao dao.Sprint,
	sprintDaoV2 daov2.Sprint,
	sprintParticipant entity.SprintParticipant,
) *UpdateSprintParticipantMutation {
	return &UpdateSprintParticipantMutation{
		dataCollector:          dataCollector,
		stateSyncer:            stateSyncer,
		sprintParticipantDao:   sprintParticipantDao,
		sprintParticipantDaoV2: sprintParticipantDaoV2,
		sprintDao:              sprintDao,
		sprintDaoV2:            sprintDaoV2,
		id:                     stateSyncer.NextMutationID(),
		sprintParticipant:      sprintParticipant,
	}
}
