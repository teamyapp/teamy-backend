package resolver

import (
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Comment struct {
	deps *Dependencies
	entity.Comment
}

func (c Comment) Commenter() (User, error) {
	userID, err := fromGraphQLID(c.CommenterID)
	if err != nil {
		return User{}, nil
	}
	user, err := c.deps.userRepo.FindUser(userID)
	if err != nil {
		return User{}, err
	}
	return newUser(c.deps, user), nil
}

func (c Comment) Task() (Task, error) {
	task, err := c.deps.Data.GetTask(c.TaskID)
	if err != nil {
		return Task{}, err
	}
	return newTask(c.deps, task), nil
}

func Comments(deps *Dependencies, cs []entity.Comment) (comments []Comment) {
	for _, c := range cs {
		comments = append(comments, Comment{
			deps:    deps,
			Comment: c,
		})
	}
	return
}
