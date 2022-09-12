package authorization

type operation string

const (
	// General operations
	createOperation operation = "Create"
	updateOperation operation = "Update"
	deleteOperation operation = "Delete"
	readOperation   operation = "Read"

	// Team specific operations
	updateSettingsOperation operation = "UpdateSettings"

	addMemberToOperation      operation = "AddMemberTo"
	removeMemberFromOperation operation = "RemoveMemberFrom"

	updateInvitationOperation operation = "UpdateInvitation"
	createInvitationOperation operation = "CreateInvitation"
	deleteInvitationOperation operation = "DeleteInvitation"
	readInvitationOperation   operation = "ReadInvitation"

	updateSprintOperation operation = "UpdateSprint"
	createSprintOperation operation = "CreateSprint"
	deleteSprintOperation operation = "DeleteSprint"
	readSprintOperation   operation = "ReadSprint"

	updateProjectOperation operation = "UpdateProject"
	createProjectOperation operation = "CreateProject"
	deleteProjectOperation operation = "DeleteProject"
	readProjectOperation   operation = "ReadProject"

	updateTaskOperation operation = "UpdateTask"
	createTaskOperation operation = "CreateTask"
	deleteTaskOperation operation = "DeleteTask"
	readTaskOperation   operation = "ReadTask"
)
