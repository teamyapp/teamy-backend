package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteTeamMember struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	teamMemberDaoV2  daov2.TeamMember
	id               uint64
	teamID           uint64
	userID           uint64
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*DeleteTeamMember)(nil)

func (d *DeleteTeamMember) GetID() uint64 {
	return d.id
}

func (d *DeleteTeamMember) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return d.teamMemberDaoV2.DeleteTeamMember(ct, tx, d.teamID, d.userID)
}

func (d *DeleteTeamMember) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if d.notifierPrepared {
		return nil
	}

	var err *errs.Error
	d.clientNotifiers, err = d.stateSyncer.GetClientNotifiersByTeamID(ct, d.teamID)
	if err != nil {
		return err
	}

	d.notifierPrepared = true
	return nil
}

func (d *DeleteTeamMember) Undo() *errs.Error {
	return nil
}

func (d *DeleteTeamMember) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return d.clientNotifiers
}

func (d *DeleteTeamMember) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.TeamMemberCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload: entity.TeamMember{
			TeamID: d.teamID,
			UserID: d.userID,
		},
	}
}

func (d *DeleteTeamMember) CleanUp(ct context.Context) *errs.Error {
	teamNotifier, err := d.stateSyncer.GetTeamNotifier(ct, d.teamID)
	if err != nil {
		d.logger.ErrorWithContext(ct, err)
		return err
	}

	teamNotifier.UnregisterUserNotifier(d.userID)
	return nil
}

func NewDeleteTeamMember(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	teamMemberDaoV2 daov2.TeamMember,
	teamID uint64,
	userID uint64,
) *DeleteTeamMember {
	return &DeleteTeamMember{
		logger:           logger,
		stateSyncer:      stateSyncer,
		teamMemberDaoV2:  teamMemberDaoV2,
		id:               stateSyncer.NextMutationID(),
		teamID:           teamID,
		userID:           userID,
		notifierPrepared: false,
	}
}
