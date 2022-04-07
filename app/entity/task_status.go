package entity

type TaskStatus string

const (
	TaskStatusNeedAttention TaskStatus = "NEED_ATTENTION"
	TaskStatusUpcoming      TaskStatus = "UPCOMING"
	TaskStatusInProgress    TaskStatus = "IN_PROGRESS"
	TaskStatusDelivered     TaskStatus = "DELIVERED"
)
