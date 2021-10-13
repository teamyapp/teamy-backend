package resolver

import (
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entity"
)

func toGraphQLTasks(tasks []entity.Task) []Task {
	gqlTasks := make([]Task, 0)
	for _, task := range tasks {
		gqlTasks = append(gqlTasks, newTask(task))
	}
	return gqlTasks
}

func toGraphQLInt(num *int) *int32 {
	if num == nil {
		return nil
	}
	gqlInt := int32(*num)
	return &gqlInt
}

func toGraphQLTime(time *time.Time) *graphql.Time {
	if time == nil {
		return nil
	}
	return &graphql.Time{Time: *time}
}
