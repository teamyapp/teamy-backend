package resolver

import (
	"github.com/teamyapp/cloud/app/api/rpc"
	"github.com/teamyapp/teamy-backend/core/collection"
	"github.com/teamyapp/teamy-backend/core/dao"
)

type Dependencies struct {
	userDao                    dao.User
	teamDao                    dao.Team
	teamMemberDao              dao.TeamMember
	invitationDao              dao.Invitation
	taskDao                    dao.Task
	threadDao                  dao.Thread
	messageDao                 dao.Message
	taskAwaitForRelationDao    dao.TaskAwaitForRelation
	userSyncer                 collection.UserSyncer
	teamSyncer                 collection.TeamSyncer
	teamMemberSyncer           collection.TeamMemberSyncer
	invitationSyncer           collection.InvitationSyncer
	taskSyncer                 collection.TaskSyncer
	threadSyncer               collection.ThreadSyncer
	messageSyncer              collection.MessageSyncer
	taskAwaitForRelationSyncer collection.TaskAwaitForRelationSyncer
	cloudAPIClient             *rpc.CloudAPIClient
}

func NewDependencies(
	userDao dao.User,
	teamDao dao.Team,
	teamMemberDao dao.TeamMember,
	invitationDao dao.Invitation,
	taskDao dao.Task,
	threadDao dao.Thread,
	messageDao dao.Message,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
	userSyncer collection.UserSyncer,
	teamSyncer collection.TeamSyncer,
	teamMemberSyncer collection.TeamMemberSyncer,
	invitationSyncer collection.InvitationSyncer,
	taskSyncer collection.TaskSyncer,
	threadSyncer collection.ThreadSyncer,
	messageSyncer collection.MessageSyncer,
	taskAwaitForRelationSyncer collection.TaskAwaitForRelationSyncer,
	cloudAPIClient *rpc.CloudAPIClient,
) *Dependencies {
	return &Dependencies{
		userDao:                    userDao,
		teamDao:                    teamDao,
		teamMemberDao:              teamMemberDao,
		invitationDao:              invitationDao,
		taskDao:                    taskDao,
		threadDao:                  threadDao,
		messageDao:                 messageDao,
		taskAwaitForRelationDao:    taskAwaitForRelationDao,
		userSyncer:                 userSyncer,
		teamSyncer:                 teamSyncer,
		teamMemberSyncer:           teamMemberSyncer,
		invitationSyncer:           invitationSyncer,
		taskSyncer:                 taskSyncer,
		threadSyncer:               threadSyncer,
		messageSyncer:              messageSyncer,
		taskAwaitForRelationSyncer: taskAwaitForRelationSyncer,
		cloudAPIClient:             cloudAPIClient,
	}
}
