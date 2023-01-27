package authorization

type ResourceType string

const (
	TeamResourceType       ResourceType = "Team"
	TaskResourceType       ResourceType = "Task"
	TaskLinkResourceType   ResourceType = "TaskLink"
	SprintResourceType     ResourceType = "Sprint"
	ProjectResourceType    ResourceType = "Project"
	InvitationResourceType ResourceType = "Invitation"
)
