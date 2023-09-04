package gql

import (
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/service"
)

type Dependencies struct {
	logger            telemetry.Logger
	taskService       service.Task
	taskLinkService   service.TaskLink
	teamService       service.Team
	sprintService     service.Sprint
	invitationService service.Invitation
	userService       service.User
	appService        service.App
	groupService      service.Group
	threadService     service.Thread
	rolloutService    service.Rollout
}

func NewDependencies(
	logger telemetry.Logger,
	taskService service.Task,
	taskLinkService service.TaskLink,
	teamService service.Team,
	sprintService service.Sprint,
	userService service.User,
	appService service.App,
	groupService service.Group,
	invitationService service.Invitation,
	threadService service.Thread,
	rolloutService service.Rollout,
) *Dependencies {
	return &Dependencies{
		logger:            logger,
		taskService:       taskService,
		taskLinkService:   taskLinkService,
		teamService:       teamService,
		sprintService:     sprintService,
		userService:       userService,
		appService:        appService,
		groupService:      groupService,
		invitationService: invitationService,
		threadService:     threadService,
		rolloutService:    rolloutService,
	}
}
