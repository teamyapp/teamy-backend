package authorization

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

var TeamAdminResourceTypeOperations = append([]ResourceTypeOperation{
	AddMemberToInTeam,
	RemoveMemberFromInTeam,
	UpdateMembersInTeam,
	ReadInvitationInTeam,
	CreateInvitationInTeam,
	DeleteInvitationInTeam,
	UpdateInvitationInTeam,
	CreateAppInstallationInTeam,
	UpdateAppInstallationInTeam,
	DeleteAppInstallationInTeam,
}, TeamMemberResourceTypeOperations...)

var TeamOwnerResourceTypeOperations = append([]ResourceTypeOperation{
	PromoteMemberToAdminInTeam,
	DemoteMemberFromAdminInTeam,
}, TeamAdminResourceTypeOperations...)

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
