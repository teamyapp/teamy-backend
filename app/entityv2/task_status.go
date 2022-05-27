package entityv2

type TaskStatus string

const (
	TaskStatusUpcoming   TaskStatus = "UPCOMING"
	TaskStatusInProgress TaskStatus = "IN_PROGRESS"
	TaskStatusDelivered  TaskStatus = "DELIVERED"
)
