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
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const invitationCodeLen = 20

func (m Mutation) CreateInvitation(ct context.Context, args struct {
	TeamID     graphql.ID
	Invitation struct {
		ReceiverFirstName *string
		ReceiverLastName  *string
		ReceiverEmail     *string
		ExpireAt          graphql.Time
	}
}) (Invitation, error) {
	senderID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	genInvitationIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "invitationID"}
	genInvitationIDRes, err := m.deps.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genInvitationIDReq)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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
	err = m.deps.invitationSyncer.CreateAndSyncInvitation(ct, invitation)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	invitation, err := m.deps.invitationDao.FindInvitationByID(ct, invitationID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	invitation.ReceiverFirstName = args.Input.ReceiverFirstName
	invitation.ReceiverLastName = args.Input.ReceiverLastName
	invitation.ExpireAt = args.Input.ExpireAt.Time
	now := time.Now()
	invitation.UpdatedAt = &now
	err = m.deps.invitationSyncer.UpdateAndSyncInvitation(ct, invitation)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	return newInvitation(m.deps, invitation), nil
}

func (m Mutation) DeleteInvitation(ct context.Context, args struct {
	InvitationID graphql.ID
}) (Invitation, error) {
	invitationID, err := fromGraphQLID(args.InvitationID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	invitation, err := m.deps.invitationDao.FindInvitationByID(ct, invitationID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	err = m.deps.invitationSyncer.DeleteAndSyncInvitation(ct, invitationID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	return newInvitation(m.deps, invitation), nil
}

func (m Mutation) AcceptInvitation(ct context.Context, args struct {
	InvitationID   graphql.ID
	InvitationCode string
}) (Invitation, error) {
	receiverUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	invitationID, err := fromGraphQLID(args.InvitationID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	invitation, err := m.deps.invitationDao.FindInvitationByID(ct, invitationID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	if invitation.Code != args.InvitationCode {
		err = errors.New("invalid invitation code")
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"invitationID":   args.InvitationID,
				"invitationCode": args.InvitationCode,
			},
		})
		return Invitation{}, err
	}

	err = m.ensureInvitationPending(ct, invitation)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	invitation.Status = entity.InvitationStatusAccepted
	invitation.ReceiverUserID = &receiverUserID
	now := time.Now()
	invitation.UpdatedAt = &now
	err = m.deps.invitationSyncer.UpdateAndSyncInvitation(ct, invitation)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	_, err = m.deps.teamMemberDao.FindTeamMember(ct, invitation.TeamID, receiverUserID)
	if err != nil {
		if !errors.As(err, &dao.ErrorNotFound) {
			m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return Invitation{}, err
		}

		_, err = m.deps.teamService.AddMemberToTeam(ct, invitation.TeamID, receiverUserID)
		if err != nil {
			m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return Invitation{}, err
		}
	}

	return newInvitation(m.deps, invitation), nil
}

func (m Mutation) DeclineInvitation(ct context.Context, args struct {
	InvitationID   graphql.ID
	InvitationCode string
}) (Invitation, error) {
	receiverUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	invitationID, err := fromGraphQLID(args.InvitationID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	invitation, err := m.deps.invitationDao.FindInvitationByID(ct, invitationID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	if invitation.Code != args.InvitationCode {
		err = errors.New("invalid invitation code")
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"InvitationID":   args.InvitationID,
				"InvitationCode": args.InvitationCode,
			},
		})
		return Invitation{}, err
	}

	err = m.ensureInvitationPending(ct, invitation)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	invitation.Status = entity.InvitationStatusDeclined
	invitation.ReceiverUserID = &receiverUserID
	now := time.Now()
	invitation.UpdatedAt = &now
	err = m.deps.invitationSyncer.UpdateAndSyncInvitation(ct, invitation)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Invitation{}, err
	}

	return newInvitation(m.deps, invitation), nil
}

func (m Mutation) ensureInvitationPending(ct context.Context, invitation entity.Invitation) error {
	switch invitation.Status {
	case entity.InvitationStatusExpired:
		err := errors.New("invitation is expired")
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"InvitationID": invitation.ID,
			},
		})
		return err
	case entity.InvitationStatusInvoked:
		err := errors.New("invitation is revoked")
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"InvitationID": invitation.ID,
			},
		})
		return err
	case entity.InvitationStatusAccepted, entity.InvitationStatusDeclined:
		err := errors.New("invitation is already responded")
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
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
