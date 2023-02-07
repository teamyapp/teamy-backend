package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateTeamMutation struct {
	dataCollector telemetry.DataCollector
	stateSyncer   *realtime.StateSyncer
	teamDao       dao.Team
	id            uint64
	team          entity.Team
}

func (c *CreateTeamMutation) GetID() uint64 {
	return c.id
}

func (c *CreateTeamMutation) Execute(ct context.Context) *errs.Error {
	err := c.teamDao.CreateTeam(ct, c.team)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return err
	}

	return nil
}

func (c *CreateTeamMutation) Undo() *errs.Error {
	return nil
}

func (c *CreateTeamMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	return c.stateSyncer.GetClientNotifiersByTeamID(ct, c.team.ID)
}

func (c *CreateTeamMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.TeamCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.team,
	}
}

func (c *CreateTeamMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateTeamMutation(
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	teamDao dao.Team,
	team entity.Team,
) *CreateTeamMutation {
	return &CreateTeamMutation{
		dataCollector: dataCollector,
		stateSyncer:   stateSyncer,
		teamDao:       teamDao,
		id:            stateSyncer.NextMutationID(),
		team:          team,
	}
}
