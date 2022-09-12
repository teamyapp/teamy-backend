package authorization

import "fmt"

type Query struct {
	ResourceType resourceType
	ResourceID   uint64
	Operation    operation
	UserID       uint64
}

func NewUpdateTeamSettingsQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: teamResourceType,
		ResourceID:   teamID,
		Operation:    updateSettingsOperation,
		UserID:       userID,
	}
}

func NewUpdateTaskQuery(userID uint64, teamID uint64) Query {
	return Query{
		ResourceType: taskResourceType,
		ResourceID:   teamID,
		Operation:    updateOperation,
		UserID:       userID,
	}
}

func (q Query) String() string {
	return fmt.Sprintf("user %d is not allowed to perform %s on %s %d", q.UserID, q.Operation, q.ResourceType, q.ResourceID)
}
