package entity

type TaskStatus int

const (
	TaskStatusNeedAttention TaskStatus = iota
	TaskStatusUpcoming
	TaskStatusInProgress
	TaskStatusDelivered
)
