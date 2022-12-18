package authorization

type ResourceTypeOperation struct {
	ResourceType ResourceType
	Operation    string
}

// Operations for Team
var (
	ReadTeam = ResourceTypeOperation{
		ResourceType: TeamResourceType,
		Operation:    "Read",
	}
	UpdateSettingsTeam = ResourceTypeOperation{
		ResourceType: TeamResourceType,
		Operation:    "UpdateSettings",
	}
	DeleteTeam = ResourceTypeOperation{
		ResourceType: TeamResourceType,
		Operation:    "Delete",
	}
	AddMemberToTeam = ResourceTypeOperation{
		ResourceType: TeamResourceType,
		Operation:    "AddMemberTo",
	}
	RemoveMemberFromTeam = ResourceTypeOperation{
		ResourceType: TeamResourceType,
		Operation:    "RemoveMemberFrom",
	}
	ReadInvitationUnderTeam = ResourceTypeOperation{
		ResourceType: TeamResourceType,
		Operation:    "ReadInvitation",
	}
	CreateInvitationUnderTeam = ResourceTypeOperation{
		ResourceType: TeamResourceType,
		Operation:    "CreateInvitation",
	}
	UpdateInvitationUnderTeam = ResourceTypeOperation{
		ResourceType: TeamResourceType,
		Operation:    "UpdateInvitation",
	}
	DeleteInvitationUnderTeam = ResourceTypeOperation{
		ResourceType: TeamResourceType,
		Operation:    "DeleteInvitation",
	}
	ReadSprintUnderTeam = ResourceTypeOperation{
		ResourceType: TeamResourceType,
		Operation:    "ReadSprint",
	}
	CreateSprintUnderTeam = ResourceTypeOperation{
		ResourceType: TeamResourceType,
		Operation:    "CreateSprint",
	}
	UpdateSprintUnderTeam = ResourceTypeOperation{
		ResourceType: TeamResourceType,
		Operation:    "UpdateSprint",
	}
	DeleteSprintUnderTeam = ResourceTypeOperation{
		ResourceType: TeamResourceType,
		Operation:    "DeleteSprint",
	}
	ReadTaskUnderTeam = ResourceTypeOperation{
		ResourceType: TeamResourceType,
		Operation:    "ReadTask",
	}
	CreateTaskUnderTeam = ResourceTypeOperation{
		ResourceType: TeamResourceType,
		Operation:    "CreateTask",
	}
	UpdateTaskUnderTeam = ResourceTypeOperation{
		ResourceType: TeamResourceType,
		Operation:    "UpdateTask",
	}
	DeleteTaskUnderTeam = ResourceTypeOperation{
		ResourceType: TeamResourceType,
		Operation:    "DeleteTask",
	}
)

// Operations for Sprint
var (
	ReadSprint = ResourceTypeOperation{
		ResourceType: SprintResourceType,
		Operation:    "Read",
	}
	UpdateSprint = ResourceTypeOperation{
		ResourceType: SprintResourceType,
		Operation:    "Update",
	}
	DeleteSprint = ResourceTypeOperation{
		ResourceType: SprintResourceType,
		Operation:    "Delete",
	}
)

// Operations for Task
var (
	ReadTask = ResourceTypeOperation{
		ResourceType: TaskResourceType,
		Operation:    "Read",
	}
	UpdateTask = ResourceTypeOperation{
		ResourceType: TaskResourceType,
		Operation:    "Update",
	}
	DeleteTask = ResourceTypeOperation{
		ResourceType: TaskResourceType,
		Operation:    "Delete",
	}
)

// Operations for Invitation
var (
	ReadInvitation = ResourceTypeOperation{
		ResourceType: InvitationResourceType,
		Operation:    "Read",
	}
	UpdateInvitation = ResourceTypeOperation{
		ResourceType: InvitationResourceType,
		Operation:    "Update",
	}
	DeleteInvitation = ResourceTypeOperation{
		ResourceType: InvitationResourceType,
		Operation:    "Delete",
	}
)

var TeamAdminResourceTypeOperations = []ResourceTypeOperation{
	AddMemberToTeam,
	RemoveMemberFromTeam,
	UpdateSettingsTeam,
	DeleteTeam,
	ReadInvitationUnderTeam,
	CreateInvitationUnderTeam,
	DeleteInvitationUnderTeam,
	UpdateInvitationUnderTeam,
}

var TeamMemberResourceTypeOperations = []ResourceTypeOperation{
	CreateSprintUnderTeam,
	CreateTaskUnderTeam,
	DeleteSprintUnderTeam,
	DeleteTaskUnderTeam,
	ReadTeam,
	ReadSprintUnderTeam,
	ReadTaskUnderTeam,
	UpdateSprint,
	UpdateTaskUnderTeam,
}
