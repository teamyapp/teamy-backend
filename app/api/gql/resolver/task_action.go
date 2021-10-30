package resolver

type TaskAction string

const (
	TaskActionStart         TaskAction = "START"
	TaskActionDelete                   = "DELETE"
	TaskActionAssignOwner              = "ASSIGN_OWNER"
	TaskActionReportBlocked            = "REPORT_BLOCKED"
	TaskActionMarkComplete             = "MARK_COMPLETE"
)
