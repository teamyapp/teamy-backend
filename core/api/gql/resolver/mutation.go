package resolver

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/ctx"
	"github.com/teamyapp/teamy-backend/core/collect"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const invitationCodeLen = 20

var invitationCodeAlphabet = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var awaitableTaskStatuses = map[entity.TaskStatus]bool{
	entity.TaskStatusInProgress: true,
	entity.TaskStatusAwaiting:   true,
}

type Mutation struct {
	deps *Dependencies
}

/* Task */

func (m Mutation) CreateTask(ct context.Context, args struct {
	TeamID graphql.ID
	Task   struct {
		Goal        string
		Context     *string
		OwnerUserID *graphql.ID
		DueAt       *graphql.Time
	}
}) (Task, error) {
	userID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		return Task{}, err
	}

	genClient := m.GeneratorClient()
	genTaskIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "taskID"}
	genTaskIDRes, err := genClient.GenerateUniqueNumber(ct, genTaskIDReq)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	owningTeamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		return Task{}, err
	}

	threadID, err := m.createThread(ct)
	if err != nil {
		return Task{}, err
	}

	ownerUserID, err := fromGraphQLIDPtr(args.Task.OwnerUserID)
	if err != nil {
		return Task{}, err
	}

	task := entity.Task{
		ID:               genTaskIDRes.UniqueNumber,
		Goal:             args.Task.Goal,
		Context:          args.Task.Context,
		Status:           entity.TaskStatusUpcoming,
		CreatorUserID:    userID,
		OwningTeamID:     owningTeamID,
		OwnerUserID:      ownerUserID,
		CommentsThreadID: threadID,
		CreatedAt:        time.Now(),
	}

	if args.Task.DueAt != nil {
		dueAt := (*args.Task.DueAt).Time
		task.DueAt = &dueAt
	}

	err = m.deps.taskSyncer.CreateAndSyncTask(task)
	if err != nil {
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) UpdateTask(ct context.Context, args struct {
	TaskID graphql.ID
	Input  struct {
		Goal         string
		Context      *string
		OwnerUserID  *graphql.ID
		OwningTeamID graphql.ID
		Effort       *int32
		DueAt        *graphql.Time
	}
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		return Task{}, err
	}

	task, err := m.deps.taskDao.FindTaskByID(taskID)
	if err != nil {
		return Task{}, err
	}

	task.Goal = args.Input.Goal
	task.Context = args.Input.Context
	ownerUserID, err := fromGraphQLIDPtr(args.Input.OwnerUserID)
	if err != nil {
		return Task{}, err
	}

	task.OwnerUserID = ownerUserID
	owningTeamID, err := fromGraphQLID(args.Input.OwningTeamID)
	if err != nil {
		return Task{}, err
	}

	task.OwningTeamID = owningTeamID
	task.Effort = intPtrFromIntPtr(args.Input.Effort)
	task.DueAt = fromGraphQLTimePtr(args.Input.DueAt)
	updatedAt := time.Now()
	task.UpdatedAt = &updatedAt
	err = m.deps.taskSyncer.UpdateAndSyncTask(task)
	if err != nil {
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) DeleteTask(ct context.Context, args struct {
	TaskID graphql.ID
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		return Task{}, err
	}

	task, err := m.deps.taskDao.FindTaskByID(taskID)
	if err != nil {
		return Task{}, err
	}

	err = m.deps.taskSyncer.DeleteAndSyncTask(taskID)
	if err != nil {
		return Task{}, err
	}

	err = m.deps.threadSyncer.DeleteAndSyncThread(task.CommentsThreadID)
	if err != nil {
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) MoveTaskToUpcoming(ct context.Context, args struct {
	TaskID graphql.ID
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		return Task{}, err
	}

	task, err := m.moveTaskToUpcoming(taskID)
	if err != nil {
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) MoveTaskToInProgress(ct context.Context, args struct {
	TaskID graphql.ID
}) (Task, error) {
	userID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		return Task{}, err
	}

	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		return Task{}, err
	}

	task, err := m.deps.taskDao.FindTaskByID(taskID)
	if err != nil {
		return Task{}, err
	}

	tasks, err := m.deps.taskDao.FindTasksByTeamID(task.OwningTeamID)
	if err != nil {

	}

	if task.OwnerUserID == nil {
		task.OwnerUserID = &userID
	}

	inProgressTasks := collect.Filter(tasks, func(eachTask entity.Task) bool {
		if eachTask.OwnerUserID == nil {
			return false
		}

		if *eachTask.OwnerUserID != *task.OwnerUserID {
			return false
		}

		return eachTask.Status == entity.TaskStatusInProgress
	})

	now := time.Now()
	if len(inProgressTasks) > 0 {
		inProgressTask := inProgressTasks[0]
		inProgressTask.Status = entity.TaskStatusPaused
		inProgressTask.UpdatedAt = &now
		err = m.deps.taskSyncer.UpdateAndSyncTask(inProgressTask)
		if err != nil {
			return Task{}, err
		}
	}

	task.Status = entity.TaskStatusInProgress
	task.UpdatedAt = &now
	err = m.deps.taskSyncer.UpdateAndSyncTask(task)
	if err != nil {
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) MoveTaskToDelivered(ct context.Context, args struct {
	TaskID graphql.ID
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		return Task{}, err
	}

	task, err := m.deps.taskDao.FindTaskByID(taskID)
	if err != nil {
		return Task{}, err
	}

	task.Status = entity.TaskStatusDelivered
	now := time.Now()
	task.UpdatedAt = &now
	err = m.deps.taskSyncer.UpdateAndSyncTask(task)
	if err != nil {
		return Task{}, err
	}

	awaitingTaskIDs, err := m.deps.taskAwaitForRelationDao.FindAwaitingTaskIDs(taskID)
	if err != nil {
		return Task{}, err
	}

	for _, awaitingTaskID := range awaitingTaskIDs {
		awaitForTaskIDs, err := m.deps.taskAwaitForRelationDao.FindAwaitForTaskIDs(awaitingTaskID)
		if err != nil {
			return Task{}, err
		}

		awaitForTasks, err := m.deps.taskDao.FindTasksByIDs(awaitForTaskIDs)
		if err != nil {
			return Task{}, err
		}

		awaitForTasks = collect.Filter(awaitForTasks, func(awaitForTask entity.Task) bool {
			return awaitForTask.Status != entity.TaskStatusDelivered
		})
		if len(awaitForTasks) == 0 {
			_, err = m.moveTaskToUpcoming(awaitingTaskID)
			if err != nil {
				return Task{}, err
			}
		}
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) moveTaskToUpcoming(taskID uint64) (entity.Task, error) {
	task, err := m.deps.taskDao.FindTaskByID(taskID)
	if err != nil {
		return entity.Task{}, err
	}

	task.Status = entity.TaskStatusUpcoming
	now := time.Now()
	task.UpdatedAt = &now
	err = m.deps.taskSyncer.UpdateAndSyncTask(task)
	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (m Mutation) MoveTaskToBlocked(ct context.Context, args struct {
	TaskID graphql.ID
	Reason string
}) (Task, error) {
	panic("implement me")
}

func (m Mutation) AddAwaitForTask(ct context.Context, args struct {
	TaskID         graphql.ID
	AwaitForTaskId graphql.ID
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		return Task{}, err
	}

	task, err := m.deps.taskDao.FindTaskByID(taskID)
	if err != nil {
		return Task{}, err
	}

	if !awaitableTaskStatuses[task.Status] {
		return Task{}, fmt.Errorf("task must be awaitable: taskID=%d", taskID)
	}

	awaitForTaskId, err := fromGraphQLID(args.AwaitForTaskId)
	if err != nil {
		return Task{}, err
	}

	now := time.Now()
	err = m.deps.taskAwaitForRelationSyncer.CreateAndSyncRelation(entity.TaskAwaitForRelation{
		AWaitingTaskID: taskID,
		AWaitForTaskID: awaitForTaskId,
		CreatedAt:      now,
	})
	if err != nil {
		return Task{}, err
	}

	task.Status = entity.TaskStatusAwaiting
	task.UpdatedAt = &now
	err = m.deps.taskSyncer.UpdateAndSyncTask(task)
	if err != nil {
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) RemoveAwaitForTask(ct context.Context, args struct {
	TaskID         graphql.ID
	AwaitForTaskId graphql.ID
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		return Task{}, err
	}

	task, err := m.deps.taskDao.FindTaskByID(taskID)
	if err != nil {
		return Task{}, err
	}

	if task.Status != entity.TaskStatusAwaiting {
		return Task{}, fmt.Errorf("task must be awaiting: taskID=%d", taskID)
	}

	awaitForTaskId, err := fromGraphQLID(args.AwaitForTaskId)
	if err != nil {
		return Task{}, err
	}

	err = m.deps.taskAwaitForRelationSyncer.DeleteAndSyncRelation(taskID, awaitForTaskId)
	if err != nil {
		return Task{}, err
	}

	awaitForTaskIds, err := m.deps.taskAwaitForRelationDao.FindAwaitForTaskIDs(taskID)
	if err != nil {
		return Task{}, err
	}

	if len(awaitForTaskIds) == 0 {
		task, err = m.moveTaskToUpcoming(taskID)
		if err != nil {
			return Task{}, err
		}
	}

	return newTask(m.deps, task), nil
}

/* Message */

func (m Mutation) CreateMessage(ct context.Context, args struct {
	ThreadID graphql.ID
	Message  struct {
		Body string
	}
}) (Message, error) {
	userID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		return Message{}, err
	}

	genClient := m.GeneratorClient()
	genMessageIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "messageID"}
	genMessageIDRes, err := genClient.GenerateUniqueNumber(ct, genMessageIDReq)
	if err != nil {
		log.Println(err)
		return Message{}, err
	}

	threadID, err := fromGraphQLID(args.ThreadID)
	if err != nil {
		return Message{}, err
	}

	message := entity.Message{
		ID:           genMessageIDRes.UniqueNumber,
		Body:         args.Message.Body,
		ThreadID:     threadID,
		AuthorUserID: userID,
		CreatedAt:    time.Now(),
	}

	err = m.deps.messageSyncer.CreateAndSyncMessage(message)
	if err != nil {
		return Message{}, err
	}

	return newMessage(m.deps, message), nil
}

func (m Mutation) UpdateMessage(ct context.Context, args struct {
	MessageID graphql.ID
	Input     struct {
		Body string
	}
}) (Message, error) {
	messageID, err := fromGraphQLID(args.MessageID)
	if err != nil {
		return Message{}, err
	}

	message, err := m.deps.messageDao.FindMessageByID(messageID)
	if err != nil {
		return Message{}, err
	}

	message.Body = args.Input.Body
	now := time.Now()
	message.UpdatedAt = &now
	err = m.deps.messageSyncer.UpdateAndSyncMessage(message)
	if err != nil {
		return Message{}, err
	}

	return newMessage(m.deps, message), nil
}

func (m Mutation) DeleteMessage(ct context.Context, args struct {
	MessageID graphql.ID
}) (Message, error) {
	messageID, err := fromGraphQLID(args.MessageID)
	if err != nil {
		return Message{}, err
	}

	message, err := m.deps.messageDao.FindMessageByID(messageID)
	if err != nil {
		return Message{}, err
	}

	err = m.deps.messageSyncer.DeleteAndSyncMessage(messageID)
	if err != nil {
		return Message{}, err
	}

	return newMessage(m.deps, message), nil
}

/* Team */

func (m Mutation) CreateTeam(ct context.Context, args struct {
	Team struct {
		Name string
	}
}) (Team, error) {
	userID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		return Team{}, err
	}

	genClient := m.GeneratorClient()
	genTeamIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "teamID"}
	genTeamIDRes, err := genClient.GenerateUniqueNumber(ct, genTeamIDReq)
	if err != nil {
		log.Println(err)
		return Team{}, err
	}

	team := entity.Team{
		ID:            genTeamIDRes.UniqueNumber,
		Name:          args.Team.Name,
		CreatorUserID: userID,
		OwnerUserID:   userID,
		CreatedAt:     time.Now(),
	}
	err = m.deps.teamSyncer.CreateAndSyncTeam(team)
	if err != nil {
		return Team{}, err
	}

	err = m.deps.teamMemberSyncer.CreateAndSyncTeamMember(team.ID, userID)
	if err != nil {
		return Team{}, err
	}

	return newTeam(m.deps, team), nil
}

func (m Mutation) UpdateTeam(ct context.Context, args struct {
	TeamID graphql.ID
	Input  struct {
		Name        string
		IconURL     *string
		OwnerUserID graphql.ID
	}
}) (Team, error) {
	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		return Team{}, err
	}

	team, err := m.deps.teamDao.FindTeamByID(teamID)
	if err != nil {
		return Team{}, err
	}

	team.Name = args.Input.Name
	team.IconURL = args.Input.IconURL
	ownerUserID, err := fromGraphQLID(args.Input.OwnerUserID)
	if err != nil {
		return Team{}, err
	}

	team.OwnerUserID = ownerUserID
	updatedAt := time.Now()
	team.UpdatedAt = &updatedAt
	err = m.deps.teamSyncer.UpdateAndSyncTeam(team)
	if err != nil {
		return Team{}, err
	}

	return newTeam(m.deps, team), err
}

func (m Mutation) AddMemberToTeam(ct context.Context, args struct {
	TeamID       graphql.ID
	MemberUserID graphql.ID
}) (User, error) {
	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	memberUserID, err := fromGraphQLID(args.MemberUserID)
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	err = m.deps.teamMemberSyncer.CreateAndSyncTeamMember(teamID, memberUserID)
	if err != nil {
		return User{}, err
	}

	user, err := m.deps.userDao.FindUserByID(memberUserID)
	if err != nil {
		return User{}, err
	}

	return newUser(m.deps, user), nil
}

func (m Mutation) RemoveMemberFromTeam(ct context.Context, args struct {
	TeamID       graphql.ID
	MemberUserID graphql.ID
}) (User, error) {
	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		return User{}, err
	}

	memberUserID, err := fromGraphQLID(args.MemberUserID)
	if err != nil {
		return User{}, err
	}

	err = m.deps.teamMemberSyncer.DeleteAndSyncTeamMember(teamID, memberUserID)
	if err != nil {
		return User{}, err
	}

	user, err := m.deps.userDao.FindUserByID(memberUserID)
	if err != nil {
		return User{}, err
	}

	return newUser(m.deps, user), nil
}

/* User */

func (m Mutation) CreateUser(ct context.Context, args struct {
	User struct {
		LastName   string
		FirstName  string
		ProfileURL *string
	}
}) (User, error) {
	userID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		return User{}, err
	}

	user := entity.User{
		ID:         userID,
		CreatedAt:  time.Now(),
		FirstName:  args.User.FirstName,
		LastName:   args.User.LastName,
		ProfileURL: args.User.ProfileURL,
	}

	err = m.deps.userSyncer.CreateAndSyncUser(user)
	if err != nil {
		return User{}, err
	}

	return newUser(m.deps, user), nil
}

func (m Mutation) UpdateUser(ct context.Context, args struct {
	UserID graphql.ID
	Input  struct {
		LastName   string
		FirstName  string
		ProfileURL *string
	}
}) (User, error) {
	userID, err := fromGraphQLID(args.UserID)
	if err != nil {
		return User{}, err
	}

	user, err := m.deps.userDao.FindUserByID(userID)
	if err != nil {
		return User{}, err
	}

	user.FirstName = args.Input.FirstName
	user.LastName = args.Input.LastName
	user.ProfileURL = args.Input.ProfileURL
	updatedAt := time.Now()
	user.UpdatedAt = &updatedAt
	err = m.deps.userSyncer.UpdateAndSyncUser(user)
	if err != nil {
		return User{}, err
	}

	return newUser(m.deps, user), nil
}

/* Invitation */

func (m Mutation) CreateInvitation(ct context.Context, args struct {
	TeamID     graphql.ID
	Invitation struct {
		ReceiverFirstName *string
		ReceiverLastName  *string
		ReceiverEmail     *string
		ExpireAt          graphql.Time
	}
}) (Invitation, error) {
	senderID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		return Invitation{}, err
	}

	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		return Invitation{}, err
	}

	genClient := m.GeneratorClient()
	genInvitationIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "invitationID"}
	genInvitationIDRes, err := genClient.GenerateUniqueNumber(ct, genInvitationIDReq)
	if err != nil {
		log.Println(err)
		return Invitation{}, err
	}

	invitation := entity.Invitation{
		ID:                genInvitationIDRes.UniqueNumber,
		SenderUserID:      senderID,
		ReceiverFirstName: args.Invitation.ReceiverFirstName,
		ReceiverLastName:  args.Invitation.ReceiverLastName,
		ReceiverEmail:     args.Invitation.ReceiverEmail,
		TeamID:            teamID,
		ExpireAt:          args.Invitation.ExpireAt.Time,
		Status:            entity.InvitationStatusPending,
		Code:              randString(invitationCodeAlphabet, invitationCodeLen),
		CreatedAt:         time.Now(),
	}
	err = m.deps.invitationSyncer.CreateAndSyncInvitation(invitation)
	if err != nil {
		return Invitation{}, err
	}

	return newInvitation(m.deps, invitation), err
}

func (m Mutation) UpdateInvitation(ct context.Context, args struct {
	InvitationID graphql.ID
	Input        struct {
		ReceiverFirstName *string
		ReceiverLastName  *string
		ExpireAt          graphql.Time
	}
}) (Invitation, error) {
	invitationID, err := fromGraphQLID(args.InvitationID)
	if err != nil {
		return Invitation{}, err
	}

	invitation, err := m.deps.invitationDao.FindInvitationByID(invitationID)
	if err != nil {
		return Invitation{}, err
	}

	invitation.ReceiverFirstName = args.Input.ReceiverFirstName
	invitation.ReceiverLastName = args.Input.ReceiverLastName
	invitation.ExpireAt = args.Input.ExpireAt.Time
	now := time.Now()
	invitation.UpdatedAt = &now
	err = m.deps.invitationSyncer.UpdateAndSyncInvitation(invitation)
	if err != nil {
		return Invitation{}, err
	}

	return newInvitation(m.deps, invitation), nil
}

func (m Mutation) DeleteInvitation(ct context.Context, args struct {
	InvitationID graphql.ID
}) (Invitation, error) {
	invitationID, err := fromGraphQLID(args.InvitationID)
	if err != nil {
		return Invitation{}, err
	}

	invitation, err := m.deps.invitationDao.FindInvitationByID(invitationID)
	if err != nil {
		return Invitation{}, err
	}

	err = m.deps.invitationSyncer.DeleteAndSyncInvitation(invitationID)
	if err != nil {
		return Invitation{}, err
	}

	return newInvitation(m.deps, invitation), nil
}

func (m Mutation) AcceptInvitation(ct context.Context, args struct {
	InvitationID   graphql.ID
	InvitationCode string
}) (Invitation, error) {
	receiverUserID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		return Invitation{}, err
	}

	invitationID, err := fromGraphQLID(args.InvitationID)
	if err != nil {
		return Invitation{}, err
	}

	invitation, err := m.deps.invitationDao.FindInvitationByID(invitationID)
	if err != nil {
		return Invitation{}, err
	}

	if invitation.Code != args.InvitationCode {
		return Invitation{}, fmt.Errorf("invalid invitation code: id=%v, code=%s\n", args.InvitationID, args.InvitationCode)
	}

	err = m.ensureInvitationPending(invitation)
	if err != nil {
		return Invitation{}, err
	}

	invitation.Status = entity.InvitationStatusAccepted
	invitation.ReceiverUserID = &receiverUserID
	now := time.Now()
	invitation.UpdatedAt = &now
	err = m.deps.invitationSyncer.UpdateAndSyncInvitation(invitation)
	if err != nil {
		return Invitation{}, err
	}

	hasMember, err := m.deps.teamMemberDao.HasTeamMember(invitation.TeamID, receiverUserID)
	if err != nil {
		return Invitation{}, err
	}

	if !hasMember {
		err = m.deps.teamMemberSyncer.CreateAndSyncTeamMember(invitation.TeamID, receiverUserID)
		if err != nil {
			return Invitation{}, err
		}
	}

	return newInvitation(m.deps, invitation), nil
}

func (m Mutation) DeclineInvitation(ct context.Context, args struct {
	InvitationID   graphql.ID
	InvitationCode string
}) (Invitation, error) {
	receiverUserID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		return Invitation{}, err
	}

	invitationID, err := fromGraphQLID(args.InvitationID)
	if err != nil {
		return Invitation{}, err
	}

	invitation, err := m.deps.invitationDao.FindInvitationByID(invitationID)
	if err != nil {
		return Invitation{}, err
	}

	if invitation.Code != args.InvitationCode {
		return Invitation{}, fmt.Errorf("invalid invitation code: id=%v, code=%s\n", args.InvitationID, args.InvitationCode)
	}

	err = m.ensureInvitationPending(invitation)
	if err != nil {
		return Invitation{}, err
	}

	invitation.Status = entity.InvitationStatusDeclined
	invitation.ReceiverUserID = &receiverUserID
	now := time.Now()
	invitation.UpdatedAt = &now
	err = m.deps.invitationSyncer.UpdateAndSyncInvitation(invitation)
	if err != nil {
		return Invitation{}, err
	}

	return newInvitation(m.deps, invitation), nil
}

func (m Mutation) ensureInvitationPending(invitation entity.Invitation) error {
	switch invitation.Status {
	case entity.InvitationStatusExpired:
		return fmt.Errorf("invitation is expired: id=%v", invitation.ID)
	case entity.InvitationStatusInvoked:
		return fmt.Errorf("invitation is revoked: id=%v", invitation.ID)
	case entity.InvitationStatusAccepted, entity.InvitationStatusDeclined:
		return fmt.Errorf("invitation is already responded: id=%v", invitation.ID)
	default:
		return nil
	}
}

func (m Mutation) createThread(ct context.Context) (uint64, error) {
	genClient := m.deps.cloudAPIClient.GeneratorClient()
	genThreadIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "threadID"}
	genThreadIDRes, err := genClient.GenerateUniqueNumber(ct, genThreadIDReq)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	threadID := genThreadIDRes.UniqueNumber
	return threadID, m.deps.threadSyncer.CreateAndSyncThread(threadID)
}

func (m Mutation) GeneratorClient() proto.GeneratorClient {
	return m.deps.cloudAPIClient.GeneratorClient()
}

func NewMutation(deps *Dependencies) Mutation {
	return Mutation{
		deps: deps,
	}
}

func randString(alphabet []rune, length int) string {
	alphabetEndIndex := len(alphabet) - 1
	result := make([]rune, length)
	for i := 0; i < length; i++ {
		randomIndex := randInt(0, alphabetEndIndex)
		result[i] = alphabet[randomIndex]
	}
	return string(result)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min+1)
}
