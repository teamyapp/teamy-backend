package resolver

type ExecutionMode struct {
}

func (e ExecutionMode) CurrUserStatus() PersonalStatus {
	panic("not implemented")
}

func (e ExecutionMode) TeamStatus() TeamStatus {
	panic("not implemented")
}
