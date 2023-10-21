package entity

type StoryAction string

const (
	StoryActionStart         StoryAction = "START"
	StoryActionDelete        StoryAction = "DELETE"
	StoryActionAssignOwner   StoryAction = "ASSIGN_OWNER"
	StoryActionReportBlocked StoryAction = "REPORT_BLOCKED"
	StoryActionMarkComplete  StoryAction = "MARK_COMPLETE"
)
