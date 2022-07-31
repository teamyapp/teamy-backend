package service

import (
	"context"
	"log"

	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Team struct {
	taskDao   dao.Task
	sprintDao dao.Sprint
	teamDao   dao.Team
}

func (t Team) FindTeams(ct context.Context, filter *TeamFilter) ([]entity.Team, error) {
	teams, err := t.teamDao.FindAllTeams()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if filter != nil {
		teams = filterTeams(teams, *filter)
	}

	return teams, nil
}

func (t Team) FindTasksInTeam(ct context.Context, teamID uint64, filter *TaskFilter) ([]entity.Task, error) {
	tasks, err := t.taskDao.FindTasksByTeamID(teamID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if filter != nil {
		tasks = filterTasks(tasks, *filter)
	}

	return tasks, nil
}

func (t Team) FindSprintsInTeam(ct context.Context, teamID uint64, filter *SprintFilter) ([]entity.Sprint, error) {
	sprints, err := t.sprintDao.FindSprintsByTeamID(teamID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if filter != nil {
		sprints = filterSprints(sprints, *filter)
	}

	return sprints, nil
}

func NewTeam(taskDao dao.Task, sprintDao dao.Sprint, teamDao dao.Team) Team {
	return Team{taskDao: taskDao, sprintDao: sprintDao, teamDao: teamDao}
}
