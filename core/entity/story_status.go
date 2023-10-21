package entity

type StoryStatus string

const (
	StoryStatusTodo       StoryStatus = "TODO"
	StoryStatusInProgress StoryStatus = "IN_PROGRESS"
	StoryStatusPaused     StoryStatus = "PAUSED"
	StoryStatusAwaiting   StoryStatus = "AWAITING"
	StoryStatusBlocked    StoryStatus = "BLOCKED"
	StoryStatusDelivered  StoryStatus = "DELIVERED"
)
