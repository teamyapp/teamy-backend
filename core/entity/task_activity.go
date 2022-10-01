package entity

type TaskActivity struct {
	TaskID           uint64
	TeamID           uint64
	DragTaskActivity DragTaskActivity
}

type DragTaskActivity struct {
	IsDragging bool
	Client     *Client
}
