package entity

type TaskActivity struct {
	IsDragging       bool
	DragByUserID     uint64
	DraggingClientID uint64
}

type TeamTaskDraggingActivity struct {
	TaskID           uint64
	TeamID           uint64
	IsDragging       bool
	DragByUserID     uint64
	DraggingClientID uint64
}
