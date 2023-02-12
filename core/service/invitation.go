package service

import (
	"context"
	"fmt"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/randgen"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

const invitationCodeLen = 20

type Invitation struct {
	dataCollector       telemetry.DataCollector
	cloudClientRegistry *cloudAPI.ClientRegistry
	authorizer          Authorizer
	stateSyncer         *realtime.StateSyncer
	invitationDao       dao.Invitation
	teamMemberDao       dao.TeamMember
	teamService         Team
}

type CreateInvitationInput struct {
	ReceiverFirstName *string
	ReceiverLastName  *string
	ReceiverEmail     *string
	ExpireAt          time.Time
}

type UpdateInvitationInput struct {
	ReceiverFirstName *string
	ReceiverLastName  *string
	ExpireAt          time.Time
}

func (i Invitation) FindInvitationsInTeam(ct context.Context, teamID uint64, filter *InvitationFilter) ([]entity.Invitation, *errs.Error) {
	invitations, err := i.invitationDao.FindInvitationsByTeamID(ct, teamID)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, err
	}

	if filter != nil {
		invitations = filterInvitations(invitations, *filter)
	}

	return invitations, nil
}

func (i Invitation) FindInvitations(ct context.Context, filter *InvitationFilter) ([]entity.Invitation, *errs.Error) {
	invitations, err := i.invitationDao.FindAllInvitations(ct)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, err
	}

	if filter != nil {
		invitations = filterInvitations(invitations, *filter)
	}

	return invitations, nil
}

func (i Invitation) CreateInvitation(ct context.Context, teamID uint64, input CreateInvitationInput) (entity.Invitation, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Invitation{}, internalErr
	}

	if feature.EnableAuthorization {
		query := authorization.NewTeamCreateInvitationQuery(userID, teamID)
		hasPermission, err := i.authorizer.hasPermission(ct, query)
		if err != nil {
			i.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.Invitation{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("permission denied: authorization query=%v", query),
			}
			i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.Invitation{}, internalErr
		}
	}

	genInvitationIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "invitationID"}
	genInvitationIDRes, rpcErr := i.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genInvitationIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Invitation{}, internalErr
	}

	invitation := entity.Invitation{
		ID:                genInvitationIDRes.UniqueNumber,
		SenderUserID:      userID,
		ReceiverFirstName: input.ReceiverFirstName,
		ReceiverLastName:  input.ReceiverLastName,
		ReceiverEmail:     input.ReceiverEmail,
		TeamID:            teamID,
		ExpireAt:          input.ExpireAt,
		Status:            entity.InvitationStatusPending,
		Code:              randgen.String(randgen.Base62, invitationCodeLen),
		CreatedAt:         time.Now(),
	}

	transaction := realtime.NewTransaction(i.dataCollector, i.stateSyncer)
	createInvitationMutation := mutation.NewCreateInvitationMutation(
		i.dataCollector,
		i.stateSyncer,
		i.invitationDao,
		invitation,
	)
	err := transaction.ApplyMutation(ct, createInvitationMutation)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Invitation{}, err
	}

	err = transaction.Commit(ct)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Invitation{}, err
	}

	if feature.EnableAuthorization {
		err = i.authorizer.registerResource(ct, authorization.InvitationResourceType, invitation.ID)
		if err != nil {
			i.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.Invitation{}, err
		}

		err = i.authorizer.assignParentResource(ct, authorization.InvitationResourceType, invitation.ID, authorization.TeamResourceType, invitation.TeamID)
		if err != nil {
			i.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.Invitation{}, err
		}
	}

	return invitation, nil
}

func (i Invitation) UpdateInvitation(ct context.Context, invitationID uint64, input UpdateInvitationInput) (entity.Invitation, *errs.Error) {
	invitation, err := i.invitationDao.FindInvitationByID(ct, invitationID)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Invitation{}, err
	}

	invitation.ReceiverFirstName = input.ReceiverFirstName
	invitation.ReceiverLastName = input.ReceiverLastName
	invitation.ExpireAt = input.ExpireAt
	now := time.Now()
	invitation.UpdatedAt = &now
	realTimeTransaction := realtime.NewTransaction(i.dataCollector, i.stateSyncer)
	updateInvitationMutation := mutation.NewUpdateInvitationMutation(
		i.dataCollector,
		i.stateSyncer,
		i.invitationDao,
		invitation,
	)
	err = realTimeTransaction.ApplyMutation(ct, updateInvitationMutation)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Invitation{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Invitation{}, err
	}

	return invitation, nil
}

func (i Invitation) DeleteInvitation(ct context.Context, invitationID uint64) (entity.Invitation, *errs.Error) {
	invitation, err := i.invitationDao.FindInvitationByID(ct, invitationID)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Invitation{}, err
	}

	realTimeTransaction := realtime.NewTransaction(i.dataCollector, i.stateSyncer)
	deleteInvitationMutation := mutation.NewDeleteInvitationMutation(
		i.dataCollector,
		i.stateSyncer,
		i.invitationDao,
		invitation,
	)

	err = realTimeTransaction.ApplyMutation(ct, deleteInvitationMutation)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Invitation{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Invitation{}, err
	}

	return invitation, nil
}

func (i Invitation) AcceptInvitation(ct context.Context, invitationID uint64, invitationCode string) (entity.Invitation, *errs.Error) {
	receiverUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Invitation{}, internalErr
	}

	invitation, err := i.invitationDao.FindInvitationByID(ct, invitationID)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Invitation{}, err
	}

	if invitation.Code != invitationCode {
		internalErr := &errs.Error{
			Code: errs.PermissionDenied,
			Message: fmt.Sprintf("invalid invitation code: invitationID=%v, invitationCode=%v",
				invitationID,
				invitationCode,
			),
		}
		i.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{
			telemetry.CauseProp: internalErr,
		})
		return entity.Invitation{}, internalErr
	}

	err = i.ensureInvitationPending(ct, invitation)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Invitation{}, err
	}

	invitation.Status = entity.InvitationStatusAccepted
	invitation.ReceiverUserID = &receiverUserID
	now := time.Now()
	invitation.UpdatedAt = &now
	realTimeTransaction := realtime.NewTransaction(i.dataCollector, i.stateSyncer)
	updateInvitationMutation := mutation.NewUpdateInvitationMutation(
		i.dataCollector,
		i.stateSyncer,
		i.invitationDao,
		invitation,
	)
	err = realTimeTransaction.ApplyMutation(ct, updateInvitationMutation)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Invitation{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Invitation{}, err
	}

	_, err = i.teamMemberDao.FindTeamMember(ct, invitation.TeamID, receiverUserID)
	if err != nil {
		if err.Code != errs.NotFound {
			i.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.Invitation{}, err
		}

		_, err = i.teamService.AddMemberToTeam(ct, invitation.TeamID, receiverUserID)
		if err != nil {
			i.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.Invitation{}, err
		}
	}

	return invitation, nil
}

func (i Invitation) DeclineInvitation(ct context.Context, invitationID uint64, invitationCode string) (entity.Invitation, *errs.Error) {
	receiverUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "user ID not found",
		}
		i.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Invitation{}, internalErr
	}

	invitation, err := i.invitationDao.FindInvitationByID(ct, invitationID)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Invitation{}, err
	}

	if invitation.Code != invitationCode {
		internalErr := &errs.Error{
			Code: errs.PermissionDenied,
			Message: fmt.Sprintf("invalid invitation code: invitationID=%v, invitationCode=%v",
				invitationID,
				invitationCode,
			),
		}
		i.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{
			telemetry.CauseProp: internalErr,
		})
		return entity.Invitation{}, internalErr
	}

	err = i.ensureInvitationPending(ct, invitation)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Invitation{}, err
	}

	invitation.Status = entity.InvitationStatusDeclined
	invitation.ReceiverUserID = &receiverUserID
	now := time.Now()
	invitation.UpdatedAt = &now
	realTimeTransaction := realtime.NewTransaction(i.dataCollector, i.stateSyncer)
	updateInvitationMutation := mutation.NewUpdateInvitationMutation(
		i.dataCollector,
		i.stateSyncer,
		i.invitationDao,
		invitation,
	)

	err = realTimeTransaction.ApplyMutation(ct, updateInvitationMutation)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Invitation{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		i.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Invitation{}, err
	}

	return invitation, nil
}

func (i Invitation) ensureInvitationPending(ct context.Context, invitation entity.Invitation) *errs.Error {
	switch invitation.Status {
	case entity.InvitationStatusExpired:
		internalErr := &errs.Error{
			Code:    errs.InvalidOperation,
			Message: fmt.Sprintf("invitation is expired: invitationID=%v", invitation.ID),
		}
		i.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{
			telemetry.CauseProp: internalErr,
		})
		return internalErr
	case entity.InvitationStatusInvoked:
		internalErr := &errs.Error{
			Code:    errs.InvalidOperation,
			Message: fmt.Sprintf("invitation is revoked: invitationID=%v", invitation.ID),
		}
		i.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{
			telemetry.CauseProp: internalErr,
		})
		return internalErr
	case entity.InvitationStatusAccepted, entity.InvitationStatusDeclined:
		internalErr := &errs.Error{
			Code:    errs.InvalidOperation,
			Message: fmt.Sprintf("invitation is already responded: invitationID=%v", invitation.ID),
		}
		i.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{
			telemetry.CauseProp: internalErr,
		})
		return internalErr
	default:
		return nil
	}
}

func NewInvitation(
	dataCollector telemetry.DataCollector,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	authorizer Authorizer,
	stateSyncer *realtime.StateSyncer,
	invitationDao dao.Invitation,
	teamMemberDao dao.TeamMember,
	teamService Team,
) Invitation {
	return Invitation{
		dataCollector:       dataCollector,
		cloudClientRegistry: cloudClientRegistry,
		authorizer:          authorizer,
		stateSyncer:         stateSyncer,
		invitationDao:       invitationDao,
		teamMemberDao:       teamMemberDao,
		teamService:         teamService,
	}
}
