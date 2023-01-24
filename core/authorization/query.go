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

func (q Query) String() string {
	return fmt.Sprintf("user %d is not allowed to perform %s on %s %d", q.UserID, q.Operation, q.ResourceType, q.ResourceID)
}
