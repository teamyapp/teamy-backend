package entity

type TaskAction string

const (
	TaskActionStart         TaskAction = "START"
	TaskActionDelete        TaskAction = "DELETE"
	TaskActionAssignOwner   TaskAction = "ASSIGN_OWNER"
	TaskActionReportBlocked TaskAction = "REPORT_BLOCKED"
	TaskActionMarkComplete  TaskAction = "MARK_COMPLETE"
)
