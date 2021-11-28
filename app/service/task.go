package service

import (
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
	taskID, err := t.taskRepo.CreateTask(task)
	if err != nil {
		log.Println(err)
		return taskID, err
	}

	activeTeam, err := t.teamRepo.GetActiveTeam(userId)
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

func (t Task) UpdateTask(task entity.Task) error {
	panic("not implemented")
}

func (t Task) StartTask(startTaskID oneEntity.ID, userID oneEntity.ID) error {
	// TODO: a user starts others' task will assign that task to the himself
	// TODO: show a modal to confirm task should be reassigned.
	activeTeam, err := t.teamRepo.GetActiveTeam(userID)
	if err != nil {
		log.Println(err)
		return err
	}

	prevNeedAttentionTaskID, err := t.taskRepo.SetNeedAttentionTask(&startTaskID, userID, activeTeam.ID)
	if err != nil {
		log.Println(err)
		return err
	}

	err = t.taskRepo.SetTeamTaskStatus(startTaskID, activeTeam.ID, entity.TaskStatusInProgress)
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

func (t Task) PerformTaskAction(taskID oneEntity.ID, action entity.TaskAction) error {
	panic("not implemented")
}

func NewTask(taskRepo repo.Task, teamRepo repo.Team) Task {
	return Task{
		taskRepo: taskRepo,
		teamRepo: teamRepo,
	}
}
