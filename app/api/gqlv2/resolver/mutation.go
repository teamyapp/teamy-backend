package resolver

type TaskInput struct {
	Goal    string
	Context *string
}

func (r Root) CreateTask(args struct{ Input TaskInput }) (Task, error) {
	task, err := r.Deps.Data.CreateTask(args.Input, 1)
	task.deps = r.Deps
	return task, err
}
