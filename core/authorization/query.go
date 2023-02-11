package authorization

import "fmt"

type Query struct {
	ResourceType ResourceType
	ResourceID   uint64
	Operation    string
	UserID       uint64
}

func NewUpdateTeamSettingsQuery(userID uint64, teamID uint64) Query {
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

func NewCreateTaskQuery(userID uint64, teamID uint64) Query {
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

func NewCreateInvitationQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: TeamResourceType,
		ResourceID:   teamID,
		Operation:    "CreateInvitation",
		UserID:       userID,
	}
}

func NewCreateSprintQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: TeamResourceType,
		ResourceID:   teamID,
		Operation:    "CreateSprint",
		UserID:       userID,
	}
}

func NewCloneTaskQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: TeamResourceType,
		ResourceID:   teamID,
		Operation:    "CloneTask",
		UserID:       userID,
	}
}

func NewUpdateTaskQuery(userID uint64, teamID uint64) Query {
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
	return fmt.Sprintf("user %d is not allowed to perform %s on %s %d", q.UserID, q.Operation, q.ResourceType, q.ResourceID)
}
