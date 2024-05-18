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
	threadService     service.Thread
	groupService      *service.Group
	rolloutService    *service.Rollout
	projectService    *service.Project
	phaseService      *service.Phase
	storyService      *service.Story
	attachmentService *service.Attachment
}

func NewDependencies(
	logger telemetry.Logger,
	taskService service.Task,
	taskLinkService service.TaskLink,
	teamService service.Team,
	sprintService service.Sprint,
	userService service.User,
	appService service.App,
	invitationService service.Invitation,
	threadService service.Thread,
	groupService *service.Group,
	rolloutService *service.Rollout,
	projectService *service.Project,
	phaseService *service.Phase,
	storyService *service.Story,
	attachmentService *service.Attachment,
) *Dependencies {
	return &Dependencies{
		logger:            logger,
		taskService:       taskService,
		taskLinkService:   taskLinkService,
		teamService:       teamService,
		sprintService:     sprintService,
		userService:       userService,
		appService:        appService,
		invitationService: invitationService,
		threadService:     threadService,
		groupService:      groupService,
		rolloutService:    rolloutService,
		projectService:    projectService,
		phaseService:      phaseService,
		storyService:      storyService,
		attachmentService: attachmentService,
	}
}
