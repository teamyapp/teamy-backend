package entityv2

type TaskStatus string

const (
	TaskStatusUpcoming TaskStatus = "UPCOMING"
	// TaskStatusInProgress requires each person has only 1 task in progress
	TaskStatusInProgress TaskStatus = "IN_PROGRESS"
	TaskStatusPaused     TaskStatus = "PAUSED"
	TaskStatusAwaiting   TaskStatus = "AWAITING"
	TaskStatusBlocked    TaskStatus = "BLOCKED"
	TaskStatusDelivered  TaskStatus = "DELIVERED"
)
