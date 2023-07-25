package authorization

import (
	"github.com/teamyapp/cloud/libs/authorization"
)

var TeamMemberResourceTypeOperations = []authorization.ResourceTypeOperation{
	ReadSprintInTeam,
	CreateSprintInTeam,
	DeleteSprintInTeam,
	ReadTaskInTeam,
	CreateTaskInTeam,
	UpdateTaskInTeam,
	DeleteTaskInTeam,
	CloneTaskInTeam,
	CreateTaskLinkInTeam,
	ReadInTeam,
	UpdateInTeam,
}

var TeamAdminResourceTypeOperations = append([]authorization.ResourceTypeOperation{
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

var TeamOwnerResourceTypeOperations = append([]authorization.ResourceTypeOperation{
	PromoteMemberToAdminInTeam,
	DemoteMemberFromAdminInTeam,
}, TeamAdminResourceTypeOperations...)

var AppMemberResourceTypeOperations = []authorization.ResourceTypeOperation{
	ReadInApp,
	CreateAppVersionInApp,
	UpdateAppVersionInApp,
	DeleteAppVersionInApp,
	CreateAppVersionVisibleTeamInApp,
	DeleteAppVersionVisibleTeamInApp,
}

var AppAdminResourceTypeOperations = append([]authorization.ResourceTypeOperation{
	UpdateInApp,
	RefreshAppSecretInApp,
	DeleteInApp,
}, AppMemberResourceTypeOperations...)

var AllTeamResourceTypeOperations = []authorization.ResourceTypeOperation{
	ReadSprintInTeam,
	CreateSprintInTeam,
	DeleteSprintInTeam,
	ReadTaskInTeam,
	CreateTaskInTeam,
	UpdateTaskInTeam,
	DeleteTaskInTeam,
	CloneTaskInTeam,
	CreateTaskLinkInTeam,
	ReadInTeam,
	UpdateInTeam,
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
}

var AllAppResourceTypeOperations = []authorization.ResourceTypeOperation{
	ReadInApp,
	UpdateInApp,
	DeleteInApp,
	RefreshAppSecretInApp,
	CreateAppVersionInApp,
	UpdateAppVersionInApp,
	DeleteAppVersionInApp,
	CreateAppVersionVisibleTeamInApp,
	DeleteAppVersionVisibleTeamInApp,
}
