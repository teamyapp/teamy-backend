package service

import (
	"context"

	"github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type Backfill struct {
	logger              telemetry.Logger
	cloudClientRegistry *api.ClientRegistry
}

func (b Backfill) BackfillPullRequestLinks(ct context.Context, teamID string) error {
	panic("implement me")
	return nil
}

func (b Backfill) BackfillParticipantsBandwidth(ct context.Context, teamID string, sprintID string) error {
	panic("implement me")
	return nil
}

func NewBackfill(
	logger telemetry.Logger,
	cloudClientRegistry *api.ClientRegistry,
) Backfill {
	return Backfill{
		logger:              logger,
		cloudClientRegistry: cloudClientRegistry,
	}
}
