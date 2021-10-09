package resolver

type TeamStatus struct {
}

func (t TeamStatus) UpcomingTasks() []Task {
	panic("not implemented")
}

func (t TeamStatus) InProgressTasks() []Task {
	panic("not implemented")
}

func (t TeamStatus) DeliveredTasks() []Task {
	panic("not implemented")
}
