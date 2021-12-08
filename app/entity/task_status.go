package entity

type TaskStatus int

const (
	TaskStatusNeedAttention TaskStatus = iota
	TaskStatusUpcoming
	TaskStatusInProgress
	TaskStatusDelivered
)

type TaskStatusEnum string

const (
	NeedAttention TaskStatusEnum = "NeedAttention"
	UPCOMING      TaskStatusEnum = "UPCOMING"
	IN_PROGRESS   TaskStatusEnum = "IN_PROGRESS"
	DELIVERED     TaskStatusEnum = "DELIVERED"
)
