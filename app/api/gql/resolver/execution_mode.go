package resolver

import (
	"github.com/teamyapp/teamy-backend/app/entity"
)

type taskActionMap = map[entity.TaskStatus][]entity.TaskAction

var availableActions = taskActionMap{
	entity.TaskStatusUpcoming: {
		entity.TaskActionStart,
		entity.TaskActionDelete,
		entity.TaskActionAssignOwner,
	},
	entity.TaskStatusNeedAttention: {
		entity.TaskActionMarkComplete,
		entity.TaskActionReportBlocked,
		entity.TaskActionAssignOwner,
		entity.TaskActionDelete,
	},
	entity.TaskStatusInProgress: {
		entity.TaskActionMarkComplete,
		entity.TaskActionReportBlocked,
		entity.TaskActionAssignOwner,
		entity.TaskActionDelete,
	},
	entity.TaskStatusDelivered: {
		entity.TaskActionDelete,
		entity.TaskActionAssignOwner,
	},
}

// func tasksWithAvailableActions(tasks []entity.Task, taskStatus entity.TaskStatus) []entity.Task {
// 	newTasks := make([]entity.Task, 0)
// 	for _, task := range tasks {
// 		newTasks = append(newTasks, taskWithAvailableActions(task, taskStatus))
// 	}
// 	return newTasks
// }

// func taskWithAvailableActions(task entity.Task, taskStatus entity.TaskStatus) entity.Task {
// 	if actions, ok := availableActions[taskStatus]; ok {
// 		task.AvailableActions = actions
// 	}
// 	return task
// }
