package authorization

type Operation string

const (
	Read   Operation = "Read"
	Create Operation = "Create"
	Update Operation = "Update"
	Delete Operation = "Delete"

	AddMemberTo      Operation = "AddMemberTo"
	RemoveMemberFrom Operation = "RemoveMemberFrom"
	UpdateSettings   Operation = "UpdateSettings"

	ReadInvitation   Operation = "ReadInvitation"
	CreateInvitation Operation = "CreateInvitation"
	UpdateInvitation Operation = "UpdateInvitation"
	DeleteInvitation Operation = "DeleteInvitation"

	ReadProject   Operation = "ReadProject"
	CreateProject Operation = "CreateProject"
	UpdateProject Operation = "UpdateProject"
	DeleteProject Operation = "DeleteProject"

	ReadSprint   Operation = "ReadSprint"
	CreateSprint Operation = "CreateSprint"
	UpdateSprint Operation = "UpdateSprint"
	DeleteSprint Operation = "DeleteSprint"

	ReadTask   Operation = "ReadTask"
	CreateTask Operation = "CreateTask"
	UpdateTask Operation = "UpdateTask"
	DeleteTask Operation = "DeleteTask"
)

var TeamAdminOperations = []Operation{
	AddMemberTo,
	RemoveMemberFrom,
	UpdateSettings,
	Delete,
	ReadInvitation,
	CreateInvitation,
	DeleteInvitation,
	UpdateInvitation,
}

var TeamMemberOperations = []Operation{
	CreateProject,
	CreateSprint,
	CreateTask,
	DeleteProject,
	DeleteSprint,
	DeleteTask,
	Read,
	ReadProject,
	ReadSprint,
	ReadTask,
	UpdateProject,
	UpdateSprint,
	UpdateTask,
}
