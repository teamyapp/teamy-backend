package entity

type TaskStatus int

const (
	TaskStatusUpcoming TaskStatus = iota
	TaskStatusInProgress
	TaskStatusDelivered
)
