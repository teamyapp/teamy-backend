package resolver

import (
	"context"
	"log"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/app/api/rpc/proto"
	"github.com/teamyapp/cloud/app/ctx"
	"github.com/teamyapp/teamy-backend/app/entity"
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

const invitationCodeLen = 8
const invitationCodeAlphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

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

	task := entityv2.Task{
		ID:               genTaskIDRes.UniqueNumber,
		Goal:             args.Task.Goal,
		Context:          args.Task.Context,
		Status:           entityv2.TaskStatusUpcoming,
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

	err = m.deps.taskDao.CreateTask(task)
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
		Status       entityv2.TaskStatus
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
	task.Status = args.Input.Status
	task.Effort = intPtrFromIntPtr(args.Input.Effort)
	task.DueAt = fromGraphQLTimePtr(args.Input.DueAt)
	updatedAt := time.Now()
	task.UpdatedAt = &updatedAt
	err = m.deps.taskDao.UpdateTask(task)
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

	err = m.deps.taskDao.DeleteTask(taskID)
	if err != nil {
		return Task{}, err
	}

	err = m.deps.threadDao.DeleteThread(task.CommentsThreadID)
	if err != nil {
		return Task{}, err
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

	message := entityv2.Message{
		ID:           genMessageIDRes.UniqueNumber,
		Body:         args.Message.Body,
		ThreadID:     threadID,
		AuthorUserID: userID,
		CreatedAt:    time.Now(),
	}

	err = m.deps.messageDao.CreateMessage(message)
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
	err = m.deps.messageDao.UpdateMessage(message)
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

	err = m.deps.messageDao.DeleteMessage(messageID)
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

	team := entityv2.Team{
		ID:        genTeamIDRes.UniqueNumber,
		Name:      args.Team.Name,
		CreatorID: userID,
		OwnerID:   userID,
		CreatedAt: time.Now(),
	}

	err = m.deps.teamDao.CreateTeam(team)
	if err != nil {
		return Team{}, err
	}

	err = m.deps.teamMemberDao.AddMemberToTeam(team.ID, userID)
	if err != nil {
		return Team{}, err
	}

	return newTeam(m.deps, team), nil
}

func (m Mutation) UpdateTeam(ct context.Context, args struct {
	TeamID graphql.ID
	Input  struct {
		Name    *string
		IconURL *string
	}
}) (Team, error) {
	panic("implement me")
}

func (m Mutation) AddMemberToTeam(ct context.Context, args struct {
	TeamID   graphql.ID
	MemberID graphql.ID
}) (Team, error) {
	panic("implement me")
}

func (m Mutation) RemoveMemberFromTeam(ct context.Context, args struct {
	TeamID   graphql.ID
	MemberID graphql.ID
}) (Team, error) {
	panic("implement me")
}

func (m Mutation) RemoveTaskFromTeam(ct context.Context, args struct {
	TeamID graphql.ID
	TaskID graphql.ID
}) (Team, error) {
	panic("implement me")
}

func (m Mutation) PromoteTeamTaskToNeedAttention(ct context.Context, args struct {
	TeamID graphql.ID
	TaskID graphql.ID
}) (Team, error) {
	panic("implement me")
}

/* User */

func (m Mutation) CreateUser(ct context.Context, args struct {
	User struct {
		LastName   string
		FirstName  string
		ProfileURL *string
	}
}) (User, error) {
	genClient := m.GeneratorClient()
	genUserIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "userID"}
	genUserIDRes, err := genClient.GenerateUniqueNumber(ct, genUserIDReq)
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	user := entityv2.User{
		ID:         genUserIDRes.UniqueNumber,
		CreatedAt:  time.Now(),
		FirstName:  args.User.FirstName,
		LastName:   args.User.LastName,
		ProfileURL: args.User.ProfileURL,
	}

	err = m.deps.userDao.CreateUser(user)
	if err != nil {
		return User{}, err
	}

	return newUser(m.deps, user), nil
}

func (m Mutation) UpdateUser(ct context.Context, args struct {
	UserID graphql.ID
	Input  struct {
		LastName   *string
		FirstName  *string
		ProfileURL *string
	}
}) (User, error) {
	panic("implement me")
}

/* Invitation */

func (m Mutation) CreateInvitation(ct context.Context, args struct {
	TeamID     graphql.ID
	Invitation struct {
		ReceiverFirstName *string
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

	genStringReq := &proto.GenerateUniqueStringRequest{
		SequenceName: "invitationCode",
		StringLength: invitationCodeLen,
		Alphabet:     invitationCodeAlphabet,
	}
	genStringRes, err := genClient.GenerateUniqueString(ct, genStringReq)
	if err != nil {
		log.Println(err)
		return Invitation{}, err
	}

	invitation := entityv2.Invitation{
		ID:                genInvitationIDRes.UniqueNumber,
		SenderUserID:      senderID,
		ReceiverFirstName: args.Invitation.ReceiverFirstName,
		ReceiverEmail:     args.Invitation.ReceiverEmail,
		TeamID:            teamID,
		ExpireAt:          args.Invitation.ExpireAt.Time,
		Status:            entityv2.InvitationStatusPending,
		Code:              genStringRes.UniqueString,
		CreatedAt:         time.Now(),
	}

	err = m.deps.invitationDao.CreateInvitation(invitation)
	if err != nil {
		return Invitation{}, err
	}

	return newInvitation(m.deps, invitation), err
}

func (m Mutation) UpdateInvitation(ct context.Context, args struct {
	InvitationID graphql.ID
	Input        struct {
		SenderUserID   *graphql.ID
		ReceiverUserID *graphql.ID
		ExpireAt       *graphql.Time
		Status         *entity.InvitationStatus
	}
}) (Invitation, error) {
	panic("implement me")
}

func (m Mutation) DeleteInvitation(ct context.Context, args struct {
	InvitationID graphql.ID
}) (Invitation, error) {
	panic("implement me")
}

func (m Mutation) AcceptInvitation(ct context.Context, args struct {
	InvitationCode string
}) (Invitation, error) {
	panic("implement me")
}

func (m Mutation) DeclineInvitation(ct context.Context, args struct {
	InvitationCode string
}) (Invitation, error) {
	panic("implement me")
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
	return threadID, m.deps.threadDao.CreateThread(threadID)
}

func (m Mutation) GeneratorClient() proto.GeneratorClient {
	return m.deps.cloudAPIClient.GeneratorClient()
}

func NewMutation(deps *Dependencies) Mutation {
	return Mutation{
		deps: deps,
	}
}
