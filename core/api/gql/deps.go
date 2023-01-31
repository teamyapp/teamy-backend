package gql

import (
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/service"
)

type Dependencies struct {
	dataCollector     telemetry.DataCollector
	taskService       service.Task
	taskLinkService   service.TaskLink
	teamService       service.Team
	sprintService     service.Sprint
	invitationService service.Invitation
	userService       service.User
	threadService     service.Thread
	appService        service.App
}

func NewDependencies(
	dataCollector telemetry.DataCollector,
	taskService service.Task,
	taskLinkService service.TaskLink,
	teamService service.Team,
	sprintService service.Sprint,
	userService service.User,
	invitationService service.Invitation,
	threadService service.Thread,
	appService service.App,
) *Dependencies {
	return &Dependencies{
		dataCollector:     dataCollector,
		taskService:       taskService,
		taskLinkService:   taskLinkService,
		teamService:       teamService,
		sprintService:     sprintService,
		userService:       userService,
		invitationService: invitationService,
		threadService:     threadService,
		appService:        appService,
	}
}
