package service

import (
	"context"

	"github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/libs/telemetry"
)

type Backfill struct {
	dataCollector       telemetry.DataCollector
	cloudClientRegistry *api.ClientRegistry
}

func (b Backfill) BackfillPullRequestLinks(ct context.Context, teamID string) error {
	return nil
}

func (b Backfill) BackfillParticipantsBandwidth(ct context.Context, teamID string, sprintID string) error {
        panic("implement me")
	return nil
}

func NewBackfill(
	dataCollector telemetry.DataCollector,
	cloudClientRegistry *api.ClientRegistry,
) Backfill {
	return Backfill{
		dataCollector:       dataCollector,
		cloudClientRegistry: cloudClientRegistry,
	}
}
