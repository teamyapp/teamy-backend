package resolver

import "context"

func (r Root) Tasks(c context.Context, args struct{ ID int32 }) ([]Task, error) {
	tasks := r.Deps.Data.GetTasks([]int32{args.ID})
	for i := range tasks {
		tasks[i].deps = r.Deps
	}
	return tasks, nil
}

func (r Root) Me() (User, error) {
	u, err := r.Deps.Data.GetUser(1)
	u.deps = r.Deps
	return u, err
}
