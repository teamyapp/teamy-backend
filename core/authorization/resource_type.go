package authorization

type ResourceType string

const (
	TaskResourceType       ResourceType = "Task"
	TaskLinkResourceType   ResourceType = "TaskLink"
	TeamResourceType       ResourceType = "Team"
	SprintResourceType     ResourceType = "Sprint"
	InvitationResourceType ResourceType = "Invitation"
	AppResourceType        ResourceType = "App"
)
