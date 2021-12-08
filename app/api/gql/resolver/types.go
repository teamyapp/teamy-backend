package resolver

import (
	"fmt"
	"strconv"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"
	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

// var gqlTaskActionMap = map[entity.TaskAction]TaskAction{
// 	entity.TaskActionStart:         TaskActionStart,
// 	entity.TaskActionDelete:        TaskActionDelete,
// 	entity.TaskActionAssignOwner:   TaskActionAssignOwner,
// 	entity.TaskActionReportBlocked: TaskActionReportBlocked,
// 	entity.TaskActionMarkComplete:  TaskActionMarkComplete,
// }

func toGraphQLTasks(deps *Dependencies, tasks []entity.Task) []Task {
	gqlTasks := make([]Task, 0)
	for _, task := range tasks {
		gqlTasks = append(gqlTasks, newTask(deps, task))
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

func toGraphQLID(id oneEntity.ID) graphql.ID {
	return graphql.ID(fmt.Sprintf("%d", int(id)))
}

func toGraphQLIDs(ids []oneEntity.ID) []graphql.ID {
	graphqlIDs := make([]graphql.ID, 0)
	for _, id := range ids {
		graphqlIDs = append(graphqlIDs, toGraphQLID(id))
	}
	return graphqlIDs
}

func toGraphQLTime(time *time.Time) *graphql.Time {
	if time == nil {
		return nil
	}
	return &graphql.Time{Time: *time}
}

// func toGraphQLTaskActions(taskActions []entity.TaskAction) []TaskAction {
// 	actions := make([]TaskAction, 0)
// 	for _, action := range taskActions {
// 		actions = append(actions, gqlTaskActionMap[action])
// 	}
// 	return actions
// }

func toGraphQLUsers(deps *Dependencies, users []entity.User) []User {
	if users == nil {
		return nil
	}

	gqlUsers := make([]User, 0)
	for _, user := range users {
		gqlUsers = append(gqlUsers, newUser(deps, user))
	}
	return gqlUsers
}

func fromGraphQLTime(graphqlTime *graphql.Time) *time.Time {
	if graphqlTime == nil {
		return nil
	}
	return &graphqlTime.Time
}

func fromGraphQLIDPtr(graphqlID *graphql.ID) (*oneEntity.ID, error) {
	if graphqlID == nil {
		return nil, nil
	}

	id, err := fromGraphQLID(*graphqlID)
	if err != nil {
		log.Error(err)
	}
	return &id, err
}

func fromGraphQLID(graphqlID graphql.ID) (oneEntity.ID, error) {
	id, err := strconv.Atoi(string(graphqlID))
	if err != nil {
		log.Error(err)
	}
	return (oneEntity.ID)(id), err
}

func fromGraphQLIDs(graphqlIDs *[]graphql.ID) ([]oneEntity.ID, error) {
	if graphqlIDs == nil || len(*graphqlIDs) == 0 {
		return nil, nil
	}

	ids := make([]oneEntity.ID, 0)
	for _, graphqlID := range *graphqlIDs {
		id, err := fromGraphQLID(graphqlID)
		if err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func fromInt32(num *int32) *int {
	if num == nil {
		return nil
	}
	intNum := int(*num)
	return &intNum
}

func fromGraphQLTaskInput(taskInput TaskInput) (entity.Task, error) {
	goal := ""
	if taskInput.Goal != nil {
		goal = *taskInput.Goal
	}

	ownerId, err := fromGraphQLIDPtr(taskInput.OwnerUserID)
	if err != nil {
		log.Error(err)
		return entity.Task{}, err
	}
	context := ""
	if taskInput.Context != nil {
		context = *taskInput.Context
	}
	task := entity.Task{
		Goal:        goal,
		DueAt:       fromGraphQLTime(taskInput.DueAt),
		Context:     context,
		OwnerUserId: ownerId,
	}
	return task, nil
}
