package service

import (
	"fmt"
	"log"

	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
	"github.com/teamyapp/teamy-backend/app/repo"
)

type Task struct {
	taskRepo repo.Task
	teamRepo repo.Team
}

func (t Task) FindTask(taskID oneEntity.ID) (entity.Task, error) {
	task, err := t.taskRepo.FindTaskByID(taskID)
	if err != nil {
		log.Println(err)
		return entity.Task{}, err
	}
	return task, nil
}

func (t Task) CreateTask(task entity.Task, userId oneEntity.ID) (oneEntity.ID, error) {
	activeTeam, err := t.teamRepo.GetActiveTeam(userId)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	if activeTeam == nil {
		return 0, fmt.Errorf("user %v does not have an active team", userId)
	}

	taskID, err := t.taskRepo.CreateTask(task)
	if err != nil {
		log.Println(err)
		return taskID, err
	}

	err = t.taskRepo.AssignTaskToTeam(taskID, activeTeam.ID, entity.TaskStatusUpcoming)

	if err != nil {
		log.Println(err)
		return taskID, err
	}
	return taskID, nil
}

func (t Task) DeleteTask(taskID oneEntity.ID, userID oneEntity.ID) error {
	// Check if taskID exists - no need, we can only delete task that we can see
	// UI:
	//		delete task from all active team view
	// Delete record from team_task table
	// TODO: move task to trash instead of completely deleting it. delete task after 7 days if in action
	// TODO: clean up the task from task dependency graph for the active team

	activeTeam, err := t.teamRepo.GetActiveTeam(userID)
	if err != nil {
		log.Println(err)
		return err
	}

	err = t.taskRepo.DeleteTeamTask(taskID, activeTeam.ID)
	if err != nil {
		log.Println(err)
	}

	err = t.taskRepo.DeleteNeedAttentionTask(taskID, userID, activeTeam.ID)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (t Task) StartTask(taskID oneEntity.ID, userID oneEntity.ID) error {
	// TODO: a user starts others' task will assign that task to the himself
	// TODO: show a modal to confirm task should be reassigned.
	activeTeam, err := t.teamRepo.GetActiveTeam(userID)
	if err != nil {
		log.Println(err)
		return err
	}

	prevNeedAttentionTaskID, err := t.taskRepo.SetNeedAttentionTask(&taskID, userID, activeTeam.ID)
	if err != nil {
		log.Println(err)
		return err
	}

	err = t.taskRepo.SetTeamTaskStatus(taskID, activeTeam.ID, entity.TaskStatusInProgress)
	if err != nil {
		log.Println(err)
		return err
	}

	if prevNeedAttentionTaskID != nil {
		err = t.taskRepo.SetTeamTaskStatus(*prevNeedAttentionTaskID, activeTeam.ID, entity.TaskStatusUpcoming)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (t Task) CompleteTask(taskID oneEntity.ID, userID oneEntity.ID) error {
	activeTeam, err := t.teamRepo.GetActiveTeam(userID)
	if err != nil {
		log.Println(err)
		return err
	}

	needAttentionTask, err := t.taskRepo.FindTaskNeedAttentionForUser(userID, activeTeam.ID)
	if err != nil {
		log.Println(err)
		return err
	}

	if needAttentionTask.ID != taskID {
		return nil
	}

	_, err = t.taskRepo.SetNeedAttentionTask(nil, userID, activeTeam.ID)
	if err != nil {
		log.Println(err)
		return err
	}

	err = t.taskRepo.SetTeamTaskStatus(taskID, activeTeam.ID, entity.TaskStatusDelivered)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (t Task) UpdateTask(task entity.Task) error {
	panic("not implemented")
}

func NewTask(taskRepo repo.Task, teamRepo repo.Team) Task {
	return Task{
		taskRepo: taskRepo,
		teamRepo: teamRepo,
	}
}
