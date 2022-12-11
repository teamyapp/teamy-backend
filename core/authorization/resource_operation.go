package authorization

type ResourceOperation struct {
	Operation    string
	ResourceType ResourceType
}

var (
	// Operations for team
	ReadTeam = ResourceOperation{
		Operation:    "Read",
		ResourceType: TeamResourceType,
	}
	UpdateSettingsTeam = ResourceOperation{
		Operation:    "UpdateSettings",
		ResourceType: TeamResourceType,
	}
	DeleteTeam = ResourceOperation{
		Operation:    "Delete",
		ResourceType: TeamResourceType,
	}
	AddMemberToTeam = ResourceOperation{
		Operation:    "AddMemberTo",
		ResourceType: TeamResourceType,
	}
	RemoveMemberFromTeam = ResourceOperation{
		Operation:    "RemoveMemberFrom",
		ResourceType: TeamResourceType,
	}
	ReadInvitationUnderTeam = ResourceOperation{
		Operation:    "ReadInvitation",
		ResourceType: TeamResourceType,
	}
	CreateInvitationUnderTeam = ResourceOperation{
		Operation:    "CreateInvitation",
		ResourceType: TeamResourceType,
	}
	UpdateInvitationUnderTeam = ResourceOperation{
		Operation:    "UpdateInvitation",
		ResourceType: TeamResourceType,
	}
	DeleteInvitationUnderTeam = ResourceOperation{
		Operation:    "DeleteInvitation",
		ResourceType: TeamResourceType,
	}
	ReadSprintUnderTeam = ResourceOperation{
		Operation:    "ReadSprint",
		ResourceType: TeamResourceType,
	}
	CreateSprintUnderTeam = ResourceOperation{
		Operation:    "CreateSprint",
		ResourceType: TeamResourceType,
	}
	UpdateSprintUnderTeam = ResourceOperation{
		Operation:    "UpdateSprint",
		ResourceType: TeamResourceType,
	}
	DeleteSprintUnderTeam = ResourceOperation{
		Operation:    "DeleteSprint",
		ResourceType: TeamResourceType,
	}
	ReadTaskUnderTeam = ResourceOperation{
		Operation:    "ReadTask",
		ResourceType: TeamResourceType,
	}
	CreateTaskUnderTeam = ResourceOperation{
		Operation:    "CreateTask",
		ResourceType: TeamResourceType,
	}
	UpdateTaskUnderTeam = ResourceOperation{
		Operation:    "UpdateTask",
		ResourceType: TeamResourceType,
	}
	DeleteTaskUnderTeam = ResourceOperation{
		Operation:    "DeleteTask",
		ResourceType: TeamResourceType,
	}

	// Operations for Sprint
	ReadSprint = ResourceOperation{
		Operation:    "Read",
		ResourceType: SprintResourceType,
	}
	UpdateSprint = ResourceOperation{
		Operation:    "Update",
		ResourceType: SprintResourceType,
	}
	DeleteSprint = ResourceOperation{
		Operation:    "Delete",
		ResourceType: SprintResourceType,
	}

	// Operations for Task
	ReadTask = ResourceOperation{
		Operation:    "Read",
		ResourceType: TaskResourceType,
	}
	UpdateTask = ResourceOperation{
		Operation:    "Update",
		ResourceType: TaskResourceType,
	}
	DeleteTask = ResourceOperation{
		Operation:    "Delete",
		ResourceType: TaskResourceType,
	}

	// Operations for Invitation
	ReadInvitation = ResourceOperation{
		Operation:    "Read",
		ResourceType: InvitationResourceType,
	}
	UpdateInvitation = ResourceOperation{
		Operation:    "Update",
		ResourceType: InvitationResourceType,
	}
	DeleteInvitation = ResourceOperation{
		Operation:    "Delete",
		ResourceType: InvitationResourceType,
	}
)

var TeamAdminOperations = []ResourceOperation{
	AddMemberToTeam,
	RemoveMemberFromTeam,
	UpdateSettingsTeam,
	DeleteTeam,
	ReadInvitationUnderTeam,
	CreateInvitationUnderTeam,
	DeleteInvitationUnderTeam,
	UpdateInvitationUnderTeam,
}

var TeamMemberOperations = []ResourceOperation{
	CreateSprintUnderTeam,
	CreateTaskUnderTeam,
	DeleteSprintUnderTeam,
	DeleteTaskUnderTeam,
	ReadTeam,
	ReadSprintUnderTeam,
	ReadTaskUnderTeam,
	UpdateSprint,
	UpdateTask,
}
