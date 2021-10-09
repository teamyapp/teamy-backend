package entity

type PersonalStatus struct {
	TaskNeedAttention *Task
	UpcomingTasks     []Task
	DeliveredTasks    []Task
}
