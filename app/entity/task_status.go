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
	UPCOMING    TaskStatusEnum = "UPCOMING"
	IN_PROGRESS TaskStatusEnum = "IN_PROGRESS"
	DELIVERED   TaskStatusEnum = "DELIVERED"
)
