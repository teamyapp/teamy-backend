package entity

type TaskStatus string

const (
	TaskStatusTodo TaskStatus = "TODO"
	// TaskStatusInProgress requires each person has only 1 task in progress
	TaskStatusInProgress TaskStatus = "IN_PROGRESS"
	TaskStatusPaused     TaskStatus = "PAUSED"
	TaskStatusAwaiting   TaskStatus = "AWAITING"
	/*
		TaskStatusBlocked
		To unblock a task
		1) Delete the task
		2) Mark task awaiting on other task
		3) Collect extra context
		4) Received help (comment) from colleague
	*/
	TaskStatusBlocked   TaskStatus = "BLOCKED"
	TaskStatusDelivered TaskStatus = "DELIVERED"
)
