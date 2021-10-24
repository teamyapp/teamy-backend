package service

import (
	"log"

	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
	"github.com/teamyapp/teamy-backend/app/repo"
)

type Execution struct {
	teamService           Team
	prioritizationService Prioritization
	taskRepo              repo.Task
}

func (e Execution) GetPersonalStatus(userID oneEntity.ID) (entity.PersonalStatus, error) {
	if err := userID.IsValid(); err != nil {
		log.Println(err)
		return entity.PersonalStatus{}, err
	}

	upcomingTasks, err := e.taskRepo.FindTasksForUser(userID, entity.TaskStatusUpcoming)
	if err != nil {
		log.Println(err)
		return entity.PersonalStatus{}, err
	}

	taskNeedAttention, err := e.taskRepo.FindTaskNeedAttentionForUser(userID)
	if err != nil {
		log.Println(err)
		return entity.PersonalStatus{}, err
	}

	deliveredTasks, err := e.taskRepo.FindTasksForUser(userID, entity.TaskStatusDelivered)
	if err != nil {
		log.Println(err)
		return entity.PersonalStatus{}, err
	}

	return entity.PersonalStatus{
		TaskNeedAttention: taskNeedAttention,
		UpcomingTasks:     e.prioritizationService.prioritizeTasks(upcomingTasks),
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

	inProgressTasks, err := e.taskRepo.FindTasksForTeam(activeTeam.ID, entity.TaskStatusInProgress)
	if err != nil {
		log.Println(err)
		return entity.TeamStatus{}, err
	}

	deliveredTasks, err := e.taskRepo.FindTasksForTeam(activeTeam.ID, entity.TaskStatusDelivered)
	if err != nil {
		log.Println(err)
		return entity.TeamStatus{}, err
	}

	return entity.TeamStatus{
		UpcomingTasks:   e.prioritizationService.prioritizeTasks(upcomingTasks),
		InProgressTasks: inProgressTasks,
		DeliveredTasks:  deliveredTasks,
	}, nil
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
