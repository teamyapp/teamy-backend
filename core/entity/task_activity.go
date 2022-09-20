package entity

type TaskActivity struct {
	TaskID           uint64
	DragTaskActivity DragTaskActivity
}

type DragTaskActivity struct {
	IsDragging       bool
	DragByUserID     uint64
	DraggingClientID uint64
}
