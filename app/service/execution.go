package service

import (
	"log"

	"github.com/teamyapp/teamy-backend/app/entity"
	"github.com/teamyapp/teamy-backend/app/repo"
)

type Execution struct {
	teamService           Team
	prioritizationService Prioritization
	taskRepo              repo.Task
}

func (e Execution) GetPersonalStatus(userID entity.ID) (entity.PersonalStatus, error) {
	panic("not implemented")
}

func (e Execution) GetActiveTeamStatus(userID entity.ID) (entity.TeamStatus, error) {
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
