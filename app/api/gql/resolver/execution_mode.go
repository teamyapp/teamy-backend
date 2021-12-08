package resolver

import (
	"github.com/teamyapp/teamy-backend/app/entity"
)

type taskActionMap = map[entity.TaskStatusEnum][]entity.TaskAction

var availableActions = taskActionMap{
	entity.UPCOMING: {
		entity.TaskActionStart,
		entity.TaskActionDelete,
		entity.TaskActionAssignOwner,
	},
	entity.NeedAttention: {
		entity.TaskActionMarkComplete,
		entity.TaskActionReportBlocked,
		entity.TaskActionAssignOwner,
		entity.TaskActionDelete,
	},
	entity.IN_PROGRESS: {
		entity.TaskActionMarkComplete,
		entity.TaskActionReportBlocked,
		entity.TaskActionAssignOwner,
		entity.TaskActionDelete,
	},
	entity.DELIVERED: {
		entity.TaskActionDelete,
	},
}

// func tasksWithAvailableActions(tasks []entity.Task, taskStatus entity.TaskStatusEnum) []entity.Task {
// 	newTasks := make([]entity.Task, 0)
// 	for _, task := range tasks {
// 		newTasks = append(newTasks, taskWithAvailableActions(task, taskStatus))
// 	}
// 	return newTasks
// }

// func taskWithAvailableActions(task entity.Task, taskStatus entity.TaskStatusEnum) entity.Task {
// 	if actions, ok := availableActions[taskStatus]; ok {
// 		task.AvailableActions = actions
// 	}
// 	return task
// }
