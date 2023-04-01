package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteTeamMemberMutation struct {
	dataCollector    telemetry.DataCollector
	stateSyncer      *realtime.StateSyncer
	teamMemberDao    dao.TeamMember
	teamMemberDaoV2  daov2.TeamMember
	id               uint64
	teamID           uint64
	userID           uint64
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*DeleteTeamMemberMutation)(nil)

func (d *DeleteTeamMemberMutation) GetID() uint64 {
	return d.id
}

func (d *DeleteTeamMemberMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return d.teamMemberDaoV2.DeleteTeamMember(ct, tx, d.teamID, d.userID)
}

func (d *DeleteTeamMemberMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
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

func (d *DeleteTeamMemberMutation) Execute(ct context.Context) *errs.Error {
	err := d.teamMemberDao.DeleteTeamMember(ct, d.teamID, d.userID)
	if err != nil {
		d.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (d *DeleteTeamMemberMutation) Undo() *errs.Error {
	return nil
}

func (d *DeleteTeamMemberMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	return d.stateSyncer.GetClientNotifiersByTeamID(ct, d.teamID)
}

func (d *DeleteTeamMemberMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return d.clientNotifiers
}

func (d *DeleteTeamMemberMutation) ToMessage() realtime.MutationMessage {
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

func (d *DeleteTeamMemberMutation) CleanUp(ct context.Context) *errs.Error {
	teamNotifier, err := d.stateSyncer.GetTeamNotifier(ct, d.teamID)
	if err != nil {
		d.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	teamNotifier.UnregisterUserNotifier(d.userID)
	return nil
}

func NewDeleteTeamMemberMutation(
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	teamMemberDao dao.TeamMember,
	teamMemberDaoV2 daov2.TeamMember,
	teamID uint64,
	userID uint64,
) *DeleteTeamMemberMutation {
	return &DeleteTeamMemberMutation{
		dataCollector:    dataCollector,
		stateSyncer:      stateSyncer,
		teamMemberDao:    teamMemberDao,
		teamMemberDaoV2:  teamMemberDaoV2,
		id:               stateSyncer.NextMutationID(),
		teamID:           teamID,
		userID:           userID,
		notifierPrepared: false,
	}
}
