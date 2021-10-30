package entity

type TaskAction int

const (
	TaskActionStart TaskAction = iota
	TaskActionDelete
	TaskActionAssignOwner
	TaskActionReportBlocked
	TaskActionMarkComplete
)
