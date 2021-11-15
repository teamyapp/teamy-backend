package entity

type TeamStatus struct {
	UpcomingTasks   []Task
	InProgressTasks []Task
	DeliveredTasks  []Task
}
