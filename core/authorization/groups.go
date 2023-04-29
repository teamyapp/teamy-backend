package authorization

import (
	"github.com/teamyapp/cloud/libs/authorization"
)

var TeamMemberResourceTypeOperations = []authorization.ResourceTypeOperation{
	CreateSprintInTeam,
	CreateTaskInTeam,
	DeleteSprintInTeam,
	DeleteTaskInTeam,
	ReadInTeam,
	UpdateInTeam,
	ReadSprintInTeam,
	ReadTaskInTeam,
	UpdateTaskInTeam,
	CloneTaskInTeam,
	CreateTaskLinkInTeam,
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
