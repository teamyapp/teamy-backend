package service

import (
	"context"
	"fmt"
	"log"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/io"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Team struct {
	cloudWebAPIExternalBaseURL string
	cloudClientRegistry        *cloudAPI.ClientRegistry
	taskDao                    dao.Task
	sprintDao                  dao.Sprint
	teamDao                    dao.Team
	teamFileUploadSessionDao   dao.TeamFileUploadSession
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

func (t Team) CreateTeamIconUploadSession(ct context.Context, teamID uint64) (uint64, error) {
	res, err := t.cloudClientRegistry.FileClient().CreateUploadSession(ct, &emptypb.Empty{})
	if err != nil {
		log.Println(err)
		return 0, err
	}

	fileUploadSession := entity.TeamFileUploadSession{
		TeamID:              teamID,
		Type:                entity.IconTeamFileUploadSessionType,
		FileUploadSessionID: res.UploadSessionId,
		IsCompleted:         false,
		CreatedAt:           time.Now(),
	}
	err = t.teamFileUploadSessionDao.CreateTeamFileUploadSession(fileUploadSession)
	return res.UploadSessionId, err
}

func (t Team) FinishTeamIconUploadSession(ct context.Context, teamID uint64, fileUploadSessionID uint64) (uint64, error) {
	iconUploadSession, err := t.teamFileUploadSessionDao.FindTeamFileUploadSessionByTeamID(
		teamID,
		entity.IconTeamFileUploadSessionType,
		fileUploadSessionID)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	if iconUploadSession.IsCompleted {
		err = fmt.Errorf("icon upload session is already completed: teamID=%v, fileUploadSessionID=%v",
			teamID, fileUploadSessionID)
		log.Println(err)
		return 0, err
	}

	now := time.Now()
	iconUploadSession.IsCompleted = true
	iconUploadSession.UpdatedAt = &now
	err = t.teamFileUploadSessionDao.UpdateTeamFileUploadSession(iconUploadSession)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	findUploadSessionReq := proto.FindUploadSessionRequest{
		UploadSessionId: fileUploadSessionID,
	}
	uploadSession, err := t.cloudClientRegistry.FileClient().FindUploadSession(ct, &findUploadSessionReq)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	team, err := t.teamDao.FindTeamByID(teamID)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	iconUrl := io.GetFileURL(t.cloudWebAPIExternalBaseURL, uploadSession.FileId)
	team.IconURL = &iconUrl
	team.UpdatedAt = &now
	return fileUploadSessionID, err
}

func NewTeam(
	cloudWebAPIExternalBaseURL string,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	taskDao dao.Task,
	sprintDao dao.Sprint,
	teamDao dao.Team,
	teamFileUploadSessionDao dao.TeamFileUploadSession,
) Team {
	return Team{
		cloudWebAPIExternalBaseURL: cloudWebAPIExternalBaseURL,
		cloudClientRegistry:        cloudClientRegistry,
		taskDao:                    taskDao,
		sprintDao:                  sprintDao,
		teamDao:                    teamDao,
		teamFileUploadSessionDao:   teamFileUploadSessionDao,
	}
}
