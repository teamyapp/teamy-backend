package entity

type StoryActivity struct {
	TaskID           uint64
	TeamID           uint64
	DragTaskActivity DragStoryActivity
}

type DragStoryActivity struct {
	IsDragging bool
	Client     *Client
}
