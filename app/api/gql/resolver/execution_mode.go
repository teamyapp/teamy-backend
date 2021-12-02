package resolver

import (
	"log"

	"github.com/teamyapp/teamy-backend/app/api/gqlv2/resolver"
	"github.com/teamyapp/teamy-backend/app/entity"

	oneEntity "github.com/teamyapp/one/entity"
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
	},
}

type ExecutionMode struct {
	deps          *Dependencies
	prototypeDeps *resolver.Dependencies
	userID        oneEntity.ID
}

func (e ExecutionMode) PersonalStatus() (PersonalStatus, error) {
	if err := e.userID.IsValid(); err != nil {
		log.Println(err)
		return PersonalStatus{}, err
	}

	activeTeam, err := e.deps.teamRepo.FindActiveTeam(e.userID)
	if err != nil {
		log.Println(err)
		return PersonalStatus{}, err
	}

	upcomingTasks, err := e.deps.taskRepo.FindTasksForUser(e.userID, activeTeam.ID, entity.TaskStatusUpcoming)
	if err != nil {
		log.Println(err)
		return PersonalStatus{}, err
	}
	upcomingTasks = e.deps.prioritizationService.PrioritizeTasks(upcomingTasks)
	upcomingTasks = tasksWithAvailableActions(upcomingTasks, entity.TaskStatusUpcoming)

	taskNeedAttention, err := e.deps.taskRepo.FindTaskNeedAttentionForUser(e.userID, activeTeam.ID)
	if err != nil {
		log.Println(err)
		return PersonalStatus{}, err
	}
	if taskNeedAttention != nil {
		task := taskWithAvailableActions(*taskNeedAttention, entity.TaskStatusNeedAttention)
		taskNeedAttention = &task
	}

	deliveredTasks, err := e.deps.taskRepo.FindTasksForUser(e.userID, activeTeam.ID, entity.TaskStatusDelivered)
	if err != nil {
		log.Println(err)
		return PersonalStatus{}, err
	}

	deliveredTasks = tasksWithAvailableActions(deliveredTasks, entity.TaskStatusDelivered)
	return PersonalStatus{
		deps:          e.deps,
		prototypeDeps: e.prototypeDeps,
		personalStatus: entity.PersonalStatus{
			TaskNeedAttention: taskNeedAttention,
			UpcomingTasks:     upcomingTasks,
			DeliveredTasks:    deliveredTasks,
		}}, nil
}

func (e ExecutionMode) TeamStatus() (TeamStatus, error) {
	if err := e.userID.IsValid(); err != nil {
		log.Println(err)
		return TeamStatus{}, err
	}

	activeTeam, err := e.deps.teamRepo.FindActiveTeam(e.userID)
	if err != nil {
		log.Println(err)
		return TeamStatus{}, err
	}

	upcomingTasks, err := e.deps.taskRepo.FindTasksForTeam(activeTeam.ID, entity.TaskStatusUpcoming)
	if err != nil {
		log.Println(err)
		return TeamStatus{}, err
	}
	upcomingTasks = e.deps.prioritizationService.PrioritizeTasks(upcomingTasks)
	upcomingTasks = tasksWithAvailableActions(upcomingTasks, entity.TaskStatusUpcoming)

	inProgressTasks, err := e.deps.taskRepo.FindTasksForTeam(activeTeam.ID, entity.TaskStatusInProgress)
	if err != nil {
		log.Println(err)
		return TeamStatus{}, err
	}
	inProgressTasks = tasksWithAvailableActions(inProgressTasks, entity.TaskStatusInProgress)

	deliveredTasks, err := e.deps.taskRepo.FindTasksForTeam(activeTeam.ID, entity.TaskStatusDelivered)
	if err != nil {
		log.Println(err)
		return TeamStatus{}, err
	}

	deliveredTasks = tasksWithAvailableActions(deliveredTasks, entity.TaskStatusDelivered)
	return TeamStatus{
		deps:          e.deps,
		prototypeDeps: e.prototypeDeps,
		teamStatus: entity.TeamStatus{
			UpcomingTasks:   upcomingTasks,
			InProgressTasks: inProgressTasks,
			DeliveredTasks:  deliveredTasks,
		},
	}, nil
}

func tasksWithAvailableActions(tasks []entity.Task, taskStatus entity.TaskStatus) []entity.Task {
	newTasks := make([]entity.Task, 0)
	for _, task := range tasks {
		newTasks = append(newTasks, taskWithAvailableActions(task, taskStatus))
	}
	return newTasks
}

func taskWithAvailableActions(task entity.Task, taskStatus entity.TaskStatus) entity.Task {
	if actions, ok := availableActions[taskStatus]; ok {
		task.AvailableActions = actions
	}
	return task
}

func newExecutionMode(deps *Dependencies, prototypeDeps *resolver.Dependencies, userID oneEntity.ID) ExecutionMode {
	return ExecutionMode{
		deps:          deps,
		prototypeDeps: prototypeDeps,
		userID:        userID,
	}
}
