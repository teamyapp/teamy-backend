package gql

import (
	"context"
	"errors"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/cloud/libs/randgen"
	"github.com/teamyapp/teamy-backend/core/api/gql/scalar"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/service"
)

const invitationCodeLen = 20

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
		IsPlanned   *bool
	}
}) (Task, error) {
	owningTeamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	ownerUserID, err := fromGraphQLIDPtr(m.deps.dataCollector, args.Task.OwnerUserID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.taskService.CreateTask(ct, owningTeamID, service.CreateTaskInput{
		Goal:        args.Task.Goal,
		Context:     args.Task.Context,
		OwnerUserID: ownerUserID,
		DueAt:       fromGraphQLTimePtr(args.Task.DueAt),
		IsPlanned:   args.Task.IsPlanned,
	})
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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
		Effort       *scalar.Duration
		DueAt        *graphql.Time
	}
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	ownerUserID, err := fromGraphQLIDPtr(m.deps.dataCollector, args.Input.OwnerUserID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	owningTeamID, err := fromGraphQLID(args.Input.OwningTeamID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.taskService.UpdateTask(ct, taskID, service.UpdateTaskInput{
		Goal:         args.Input.Goal,
		Context:      args.Input.Context,
		OwnerUserID:  ownerUserID,
		OwningTeamID: owningTeamID,
		Effort:       fromGraphQLDurationPtr(args.Input.Effort),
		DueAt:        fromGraphQLTimePtr(args.Input.DueAt),
	})
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) DeleteTask(ct context.Context, args struct {
	TaskID graphql.ID
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.taskService.DeleteTask(ct, taskID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) MoveTaskToUpcoming(ct context.Context, args struct {
	TaskID graphql.ID
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.taskService.MoveTaskToUpcoming(ct, taskID, true)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) MoveTaskToInProgress(ct context.Context, args struct {
	TaskID graphql.ID
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.taskService.MoveTaskToInProgress(ct, taskID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) MoveTaskToDelivered(ct context.Context, args struct {
	TaskID graphql.ID
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.taskService.MoveTaskToDelivered(ct, taskID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) MoveTaskToBlocked(ct context.Context, args struct {
	TaskID graphql.ID
	Reason string
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.taskService.MoveTaskToBlocked(ct, taskID, args.Reason)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) AddAwaitForTask(ct context.Context, args struct {
	TaskID         graphql.ID
	AwaitForTaskId graphql.ID
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	awaitForTaskId, err := fromGraphQLID(args.AwaitForTaskId)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.taskService.AddAwaitForTask(ct, taskID, awaitForTaskId)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	awaitForTaskId, err := fromGraphQLID(args.AwaitForTaskId)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.taskService.RemoveAwaitForTask(ct, taskID, awaitForTaskId)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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
	userID, err := ctx.UserIDFromContext(m.deps.dataCollector, ct)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	genMessageIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "messageID"}
	genMessageIDRes, err := m.deps.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genMessageIDReq)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	threadID, err := fromGraphQLID(args.ThreadID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	message, err := m.deps.messageDao.FindMessageByID(messageID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	message.Body = args.Input.Body
	now := time.Now()
	message.UpdatedAt = &now
	err = m.deps.messageSyncer.UpdateAndSyncMessage(message)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	return newMessage(m.deps, message), nil
}

func (m Mutation) DeleteMessage(ct context.Context, args struct {
	MessageID graphql.ID
}) (Message, error) {
	messageID, err := fromGraphQLID(args.MessageID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	message, err := m.deps.messageDao.FindMessageByID(messageID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	err = m.deps.messageSyncer.DeleteAndSyncMessage(messageID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	return newMessage(m.deps, message), nil
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
	senderID, err := ctx.UserIDFromContext(m.deps.dataCollector, ct)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	genInvitationIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "invitationID"}
	genInvitationIDRes, err := m.deps.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genInvitationIDReq)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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
		Code:              randgen.String(randgen.Base62, invitationCodeLen),
		CreatedAt:         time.Now(),
	}
	err = m.deps.invitationSyncer.CreateAndSyncInvitation(invitation)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	invitation, err := m.deps.invitationDao.FindInvitationByID(invitationID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	invitation.ReceiverFirstName = args.Input.ReceiverFirstName
	invitation.ReceiverLastName = args.Input.ReceiverLastName
	invitation.ExpireAt = args.Input.ExpireAt.Time
	now := time.Now()
	invitation.UpdatedAt = &now
	err = m.deps.invitationSyncer.UpdateAndSyncInvitation(invitation)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	return newInvitation(m.deps, invitation), nil
}

func (m Mutation) DeleteInvitation(ct context.Context, args struct {
	InvitationID graphql.ID
}) (Invitation, error) {
	invitationID, err := fromGraphQLID(args.InvitationID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	invitation, err := m.deps.invitationDao.FindInvitationByID(invitationID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	err = m.deps.invitationSyncer.DeleteAndSyncInvitation(invitationID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	return newInvitation(m.deps, invitation), nil
}

func (m Mutation) AcceptInvitation(ct context.Context, args struct {
	InvitationID   graphql.ID
	InvitationCode string
}) (Invitation, error) {
	receiverUserID, err := ctx.UserIDFromContext(m.deps.dataCollector, ct)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	invitationID, err := fromGraphQLID(args.InvitationID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	invitation, err := m.deps.invitationDao.FindInvitationByID(invitationID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	if invitation.Code != args.InvitationCode {
		err = errors.New("invalid invitation code")
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"invitationID":   args.InvitationID,
				"invitationCode": args.InvitationCode,
			},
		})
		return Invitation{}, err
	}

	err = m.ensureInvitationPending(invitation)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	invitation.Status = entity.InvitationStatusAccepted
	invitation.ReceiverUserID = &receiverUserID
	now := time.Now()
	invitation.UpdatedAt = &now
	err = m.deps.invitationSyncer.UpdateAndSyncInvitation(invitation)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	hasMember, err := m.deps.teamMemberDao.HasTeamMember(invitation.TeamID, receiverUserID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	if !hasMember {
		err = m.deps.teamMemberSyncer.CreateAndSyncTeamMember(invitation.TeamID, receiverUserID)
		if err != nil {
			m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return Invitation{}, err
		}
	}

	return newInvitation(m.deps, invitation), nil
}

func (m Mutation) DeclineInvitation(ct context.Context, args struct {
	InvitationID   graphql.ID
	InvitationCode string
}) (Invitation, error) {
	receiverUserID, err := ctx.UserIDFromContext(m.deps.dataCollector, ct)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	invitationID, err := fromGraphQLID(args.InvitationID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	invitation, err := m.deps.invitationDao.FindInvitationByID(invitationID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	if invitation.Code != args.InvitationCode {
		err = errors.New("invalid invitation code")
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"InvitationID":   args.InvitationID,
				"InvitationCode": args.InvitationCode,
			},
		})
		return Invitation{}, err
	}

	err = m.ensureInvitationPending(invitation)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	invitation.Status = entity.InvitationStatusDeclined
	invitation.ReceiverUserID = &receiverUserID
	now := time.Now()
	invitation.UpdatedAt = &now
	err = m.deps.invitationSyncer.UpdateAndSyncInvitation(invitation)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	return newInvitation(m.deps, invitation), nil
}

func (m Mutation) ensureInvitationPending(invitation entity.Invitation) error {
	switch invitation.Status {
	case entity.InvitationStatusExpired:
		err := errors.New("invitation is expired")
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"InvitationID": invitation.ID,
			},
		})
		return err
	case entity.InvitationStatusInvoked:
		err := errors.New("invitation is revoked")
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"InvitationID": invitation.ID,
			},
		})
		return err
	case entity.InvitationStatusAccepted, entity.InvitationStatusDeclined:
		err := errors.New("invitation is already responded")
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"InvitationID": invitation.ID,
			},
		})
		return err
	default:
		return nil
	}
}

func NewMutation(deps *Dependencies) Mutation {
	return Mutation{
		deps: deps,
	}
}
