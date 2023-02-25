package authorization

import "fmt"

type Query struct {
	ResourceType ResourceType
	ResourceID   uint64
	Operation    string
	UserID       uint64
}

func NewReadTeamQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: TeamResourceType,
		ResourceID:   teamID,
		Operation:    "Read",
		UserID:       userID,
	}
}

func NewTeamUpdateSettingsQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: TeamResourceType,
		ResourceID:   teamID,
		Operation:    "UpdateSettings",
		UserID:       userID,
	}
}

func NewDeleteTeamQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: TeamResourceType,
		ResourceID:   teamID,
		Operation:    "Delete",
		UserID:       userID,
	}
}

func NewTeamAddMemberToQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: TeamResourceType,
		ResourceID:   teamID,
		Operation:    "AddMemberTo",
		UserID:       userID,
	}
}

func NewTeamRemoveMemberFromQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: TeamResourceType,
		ResourceID:   teamID,
		Operation:    "RemoveMemberFrom",
		UserID:       userID,
	}
}

func NewTeamUpdateMemberQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: TeamResourceType,
		ResourceID:   teamID,
		Operation:    "UpdateMember",
		UserID:       userID,
	}
}

func NewTeamReadMemberQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: TeamResourceType,
		ResourceID:   teamID,
		Operation:    "ReadMember",
		UserID:       userID,
	}
}

func NewTeamCreateTaskQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: TeamResourceType,
		ResourceID:   teamID,
		Operation:    "CreateTask",
		UserID:       userID,
	}
}

func NewCreateTaskLinkQuery(userID uint64, taskID uint64) Query {
	return Query{
		ResourceType: TaskResourceType,
		ResourceID:   taskID,
		Operation:    "CreateLink",
		UserID:       userID,
	}
}

func NewTeamCreateInvitationQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: TeamResourceType,
		ResourceID:   teamID,
		Operation:    "CreateInvitation",
		UserID:       userID,
	}
}

func NewTeamCreateSprintQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: TeamResourceType,
		ResourceID:   teamID,
		Operation:    "CreateSprint",
		UserID:       userID,
	}
}

func NewTeamReadSprintQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: TeamResourceType,
		ResourceID:   teamID,
		Operation:    "ReadSprint",
		UserID:       userID,
	}
}

func NewTeamCloneTaskQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: TeamResourceType,
		ResourceID:   teamID,
		Operation:    "CloneTask",
		UserID:       userID,
	}
}

func NewTeamUpdateTaskQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: TaskResourceType,
		ResourceID:   teamID,
		Operation:    "Update",
		UserID:       userID,
	}
}

func NewCreateAppTeamInstallationQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: TeamResourceType,
		ResourceID:   teamID,
		Operation:    "CreateAppInstallation",
		UserID:       userID,
	}
}

func NewUpdateAppTeamInstallationQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: TeamResourceType,
		ResourceID:   teamID,
		Operation:    "UpdateAppInstallation",
		UserID:       userID,
	}
}

func NewDeleteAppTeamInstallationQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: TeamResourceType,
		ResourceID:   teamID,
		Operation:    "DeleteAppInstallation",
		UserID:       userID,
	}
}

func NewUpdateAppQuery(userID uint64, appID uint64) Query {
	return Query{
		ResourceType: AppResourceType,
		ResourceID:   appID,
		Operation:    "Update",
		UserID:       userID,
	}
}

func NewRefreshAppSecretQuery(userID uint64, appID uint64) Query {
	return Query{
		ResourceType: AppResourceType,
		ResourceID:   appID,
		Operation:    "RefreshSecret",
		UserID:       userID,
	}
}

func NewDeleteAppQuery(userID uint64, appID uint64) Query {
	return Query{
		ResourceType: AppResourceType,
		ResourceID:   appID,
		Operation:    "Delete",
		UserID:       userID,
	}
}

func NewCreateAppVersionQuery(userID uint64, appID uint64) Query {
	return Query{
		ResourceType: AppResourceType,
		ResourceID:   appID,
		Operation:    "CreateAppVersion",
		UserID:       userID,
	}
}

func NewUpdateAppVersionQuery(userID uint64, appID uint64) Query {
	return Query{
		ResourceType: AppResourceType,
		ResourceID:   appID,
		Operation:    "UpdateAppVersion",
		UserID:       userID,
	}
}

func NewDeleteAppVersionQuery(userID uint64, appID uint64) Query {
	return Query{
		ResourceType: AppResourceType,
		ResourceID:   appID,
		Operation:    "DeleteAppVersion",
		UserID:       userID,
	}
}

func NewCreateAppVersionVisibleTeamQuery(userID uint64, appID uint64) Query {
	return Query{
		ResourceType: AppResourceType,
		ResourceID:   appID,
		Operation:    "CreateAppVersionVisibleTeam",
		UserID:       userID,
	}
}

func NewDeleteAppVersionVisibleTeamQuery(userID uint64, appID uint64) Query {
	return Query{
		ResourceType: AppResourceType,
		ResourceID:   appID,
		Operation:    "DeleteAppVersionVisibleTeam",
		UserID:       userID,
	}
}

func (q Query) String() string {
	return fmt.Sprintf("[Query UserID=%v Operation=%v ResourceType=%v ResourceID=%v]", q.UserID, q.Operation, q.ResourceType, q.ResourceID)
}
