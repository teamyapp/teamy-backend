package service

import (
	"log"

	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
	"github.com/teamyapp/teamy-backend/app/repo"
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

type Execution struct {
	teamService           Team
	prioritizationService Prioritization
	taskRepo              repo.Task
}

func (e Execution) GetActiveTeam(userID oneEntity.ID) (*entity.Team, error) {
	activeTeam, err := e.teamService.GetActiveTeam(userID)
	if err != nil {
		log.Println(err)
	}
	return activeTeam, nil
}

func (e Execution) GetPersonalStatusForActiveTeam(userID oneEntity.ID) (entity.PersonalStatus, error) {
	if err := userID.IsValid(); err != nil {
		log.Println(err)
		return entity.PersonalStatus{}, err
	}

	activeTeam, err := e.teamService.GetActiveTeam(userID)
	if err != nil {
		log.Println(err)
		return entity.PersonalStatus{}, err
	}

	upcomingTasks, err := e.taskRepo.FindTasksForUser(userID, activeTeam.ID, entity.TaskStatusUpcoming)
	if err != nil {
		log.Println(err)
		return entity.PersonalStatus{}, err
	}
	upcomingTasks = e.prioritizationService.prioritizeTasks(upcomingTasks)
	upcomingTasks = tasksWithAvailableActions(upcomingTasks, entity.TaskStatusUpcoming)

	taskNeedAttention, err := e.taskRepo.FindTaskNeedAttentionForUser(userID, activeTeam.ID)
	if err != nil {
		log.Println(err)
		return entity.PersonalStatus{}, err
	}
	if taskNeedAttention != nil {
		task := taskWithAvailableActions(*taskNeedAttention, entity.TaskStatusNeedAttention)
		taskNeedAttention = &task
	}

	deliveredTasks, err := e.taskRepo.FindTasksForUser(userID, activeTeam.ID, entity.TaskStatusDelivered)
	if err != nil {
		log.Println(err)
		return entity.PersonalStatus{}, err
	}
	deliveredTasks = tasksWithAvailableActions(deliveredTasks, entity.TaskStatusDelivered)

	return entity.PersonalStatus{
		TaskNeedAttention: taskNeedAttention,
		UpcomingTasks:     upcomingTasks,
		DeliveredTasks:    deliveredTasks,
	}, nil
}

func (e Execution) GetActiveTeamStatus(userID oneEntity.ID) (entity.TeamStatus, error) {
	if err := userID.IsValid(); err != nil {
		log.Println(err)
		return entity.TeamStatus{}, err
	}

	activeTeam, err := e.teamService.GetActiveTeam(userID)
	if err != nil {
		log.Println(err)
		return entity.TeamStatus{}, err
	}

	upcomingTasks, err := e.taskRepo.FindTasksForTeam(activeTeam.ID, entity.TaskStatusUpcoming)
	if err != nil {
		log.Println(err)
		return entity.TeamStatus{}, err
	}
	upcomingTasks = e.prioritizationService.prioritizeTasks(upcomingTasks)
	upcomingTasks = tasksWithAvailableActions(upcomingTasks, entity.TaskStatusUpcoming)

	inProgressTasks, err := e.taskRepo.FindTasksForTeam(activeTeam.ID, entity.TaskStatusInProgress)
	if err != nil {
		log.Println(err)
		return entity.TeamStatus{}, err
	}
	inProgressTasks = tasksWithAvailableActions(inProgressTasks, entity.TaskStatusInProgress)

	deliveredTasks, err := e.taskRepo.FindTasksForTeam(activeTeam.ID, entity.TaskStatusDelivered)
	if err != nil {
		log.Println(err)
		return entity.TeamStatus{}, err
	}
	deliveredTasks = tasksWithAvailableActions(deliveredTasks, entity.TaskStatusDelivered)

	return entity.TeamStatus{
		UpcomingTasks:   upcomingTasks,
		InProgressTasks: inProgressTasks,
		DeliveredTasks:  deliveredTasks,
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

func NewExecution(
	teamService Team,
	prioritizationService Prioritization,
	taskRepo repo.Task,
) Execution {
	return Execution{
		teamService:           teamService,
		prioritizationService: prioritizationService,
		taskRepo:              taskRepo,
	}
}
