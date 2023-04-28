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

type UpdateTeamMember struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	teamMemberDaoV2  daov2.TeamMember
	id               uint64
	teamMember       entity.TeamMember
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*UpdateTeamMember)(nil)

func (u *UpdateTeamMember) GetID() uint64 {
	return u.id
}

func (u *UpdateTeamMember) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return u.teamMemberDaoV2.UpdateTeamMember(ct, tx, u.teamMember)
}

func (u *UpdateTeamMember) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if u.notifierPrepared {
		return nil
	}

	var err *errs.Error
	u.clientNotifiers, err = u.stateSyncer.GetClientNotifiersByTeamID(ct, u.teamMember.TeamID)
	if err != nil {
		return err
	}

	u.notifierPrepared = true
	return nil
}

func (u *UpdateTeamMember) Undo() *errs.Error {
	return nil
}

func (u *UpdateTeamMember) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return u.clientNotifiers
}

func (u *UpdateTeamMember) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.TeamMemberCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.teamMember,
	}
}

func (u *UpdateTeamMember) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewUpdateTeamMember(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	teamMemberDaoV2 daov2.TeamMember,
	teamMember entity.TeamMember,
) *UpdateTeamMember {
	return &UpdateTeamMember{
		logger:           logger,
		stateSyncer:      stateSyncer,
		teamMemberDaoV2:  teamMemberDaoV2,
		id:               stateSyncer.NextMutationID(),
		teamMember:       teamMember,
		notifierPrepared: false,
	}
}
