package resolver

type PersonalStatus struct {
}

func (p PersonalStatus) NeedAttention() *Task {
	panic("not implemented")
}

func (p PersonalStatus) UpcomingTasks() []Task {
	panic("not implemented")
}

func (p PersonalStatus) DeliveredTasks() []Task {
	panic("not implemented")
}
