package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	daoEntity "github.com/teamyapp/teamy-backend/core/dao/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateTeamMemberGroup struct {
	logger             telemetry.Logger
	stateSyncer        *realtime.StateSyncer
	teamMemberGroupDao dao.TeamMemberGroup
	id                 uint64
	teamMemberGroup    daoEntity.TeamMemberGroup
	clientNotifiers    []*realtime.ClientNotifier
	notifierPrepared   bool
}

var _ realtime.Mutation = (*UpdateTeamMemberGroup)(nil)

func (u *UpdateTeamMemberGroup) GetID() uint64 {
	return u.id
}

func (u *UpdateTeamMemberGroup) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return u.teamMemberGroupDao.UpdateMemberGroup(ct, tx, u.teamMemberGroup)
}

func (u *UpdateTeamMemberGroup) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if u.notifierPrepared {
		return nil
	}

	var err *errs.Error
	u.clientNotifiers, err = u.stateSyncer.GetClientNotifiersByTeamID(ct, u.teamMemberGroup.TeamID)
	if err != nil {
		return err
	}

	u.notifierPrepared = true
	return nil
}

func (u *UpdateTeamMemberGroup) Undo() *errs.Error {
	return nil
}

func (u *UpdateTeamMemberGroup) GetClientNotifiers() []*realtime.ClientNotifier {
	return u.clientNotifiers
}

func (u *UpdateTeamMemberGroup) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.TeamMemberGroupCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.teamMemberGroup,
	}
}

func (u *UpdateTeamMemberGroup) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewUpdateTeamMemberGroup(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	teamMemberGroupDao dao.TeamMemberGroup,
	teamMemberGroup daoEntity.TeamMemberGroup,
) *UpdateTeamMemberGroup {
	return &UpdateTeamMemberGroup{
		logger:             logger,
		stateSyncer:        stateSyncer,
		teamMemberGroupDao: teamMemberGroupDao,
		id:                 stateSyncer.NextMutationID(),
		teamMemberGroup:    teamMemberGroup,
	}
}
