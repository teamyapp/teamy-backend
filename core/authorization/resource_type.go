package authorization

type resourceType string

const (
	teamResourceType       resourceType = "Team"
	taskResourceType       resourceType = "Task"
	sprintResourceType     resourceType = "Sprint"
	projectResourceType    resourceType = "Project"
	invitationResourceType resourceType = "Invitation"
)
