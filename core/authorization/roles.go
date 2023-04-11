package authorization

var TeamAdminResourceTypeOperations = []ResourceTypeOperation{
	AddMemberToInTeam,
	RemoveMemberFromInTeam,
	UpdateMembersInTeam,
	UpdateInTeam,
	DeleteInTeam,
	ReadInvitationInTeam,
	CreateInvitationInTeam,
	DeleteInvitationInTeam,
	UpdateInvitationInTeam,
	CreateAppInstallationInTeam,
	UpdateAppInstallationInTeam,
	DeleteAppInstallationInTeam,
}

var TeamMemberResourceTypeOperations = []ResourceTypeOperation{
	CreateSprintInTeam,
	CreateTaskInTeam,
	DeleteSprintInTeam,
	DeleteTaskInTeam,
	ReadInTeam,
	ReadSprintInTeam,
	ReadTaskInTeam,
	UpdateTaskInTeam,
	CloneTaskInTeam,
}

var AppAdminResourceTypeOperations = []ResourceTypeOperation{
	UpdateInApp,
	RefreshAppSecretInApp,
	DeleteInApp,
}

var AppMemberResourceTypeOperations = []ResourceTypeOperation{
	ReadInApp,
	CreateAppVersionInApp,
	UpdateAppVersionInApp,
	DeleteAppVersionInApp,
	CreateAppVersionVisibleTeamInApp,
	DeleteAppVersionVisibleTeamInApp,
}
