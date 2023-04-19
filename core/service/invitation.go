package service

import (
	"context"
	"fmt"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/randgen"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

const invitationCodeLen = 20

type Invitation struct {
	logger                 telemetry.Logger
	cloudClientRegistry    *cloudAPI.ClientRegistry
	authorizer             Authorizer
	featureToggles         feature.Toggles
	stateSyncer            *realtime.StateSyncer
	transactionFactory     transaction.Factory
	invitationDao          dao.Invitation
	invitationDaoV2        daov2.Invitation
	teamMemberDao          dao.TeamMember
	teamMemberDaoV2        daov2.TeamMember
	sprintParticipantDao   dao.SprintParticipant
	sprintParticipantDaoV2 daov2.SprintParticipant
	sprintDao              dao.Sprint
	sprintDaoV2            daov2.Sprint
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
	invitations, err := i.invitationDaoV2.FindInvitationsByTeamID(ct, teamID)
	if err != nil {
		return nil, err
	}

	if filter != nil {
		invitations = filterInvitations(invitations, *filter)
	}

	return invitations, nil
}

func (i Invitation) FindInvitations(ct context.Context, filter *InvitationFilter) ([]entity.Invitation, *errs.Error) {
	invitations, err := i.invitationDaoV2.FindAllInvitations(ct)
	if err != nil {
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
		return entity.Invitation{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if i.featureToggles.EnableAuthorization {
		query := authorization.NewCreateInvitationInTeamQuery(userID, teamID)
		hasPermission, err := i.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.Invitation{}, err
		}

		if !hasPermission {
			return entity.Invitation{}, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	genInvitationIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "invitationID"}
	genInvitationIDRes, rpcErr := i.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genInvitationIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
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
		CreatedAt:         time.Now().UTC(),
	}

	txCtx := TransactionsContext{
		logger:             i.logger,
		transactionFactory: i.transactionFactory,
		stateSyncer:        i.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		createInvitationMutation := mutation.NewCreateInvitation(
			i.logger,
			i.stateSyncer,
			i.invitationDao,
			i.invitationDaoV2,
			invitation,
		)

		internalErr := createInvitationMutation.ExecuteV2(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		rtTx.AppendMutation(createInvitationMutation)
		return nil
	})

	if err != nil {
		return entity.Invitation{}, err
	}

	if i.featureToggles.EnableAuthorization {
		err = i.authorizer.registerResource(ct, authorization.InvitationResourceType, invitation.ID)
		if err != nil {
			return entity.Invitation{}, err
		}

		err = i.authorizer.assignParentResource(ct, authorization.InvitationResourceType, invitation.ID, authorization.TeamResourceType, invitation.TeamID)
		if err != nil {
			return entity.Invitation{}, err
		}
	}

	return invitation, nil
}

func (i Invitation) UpdateInvitation(ct context.Context, invitationID uint64, input UpdateInvitationInput) (entity.Invitation, *errs.Error) {
	invitation, err := i.invitationDaoV2.FindInvitationByID(ct, invitationID)
	if err != nil {
		return entity.Invitation{}, err
	}

	invitation.ReceiverFirstName = input.ReceiverFirstName
	invitation.ReceiverLastName = input.ReceiverLastName
	invitation.ExpireAt = input.ExpireAt
	now := time.Now().UTC()
	invitation.UpdatedAt = &now

	txCtx := TransactionsContext{
		logger:             i.logger,
		transactionFactory: i.transactionFactory,
		stateSyncer:        i.stateSyncer,
		ct:                 ct,
	}
	err = txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		updateInvitationMutation := mutation.NewUpdateInvitation(
			i.logger,
			i.stateSyncer,
			i.invitationDao,
			i.invitationDaoV2,
			invitation,
		)
		internalErr := updateInvitationMutation.ExecuteV2(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		rtTx.AppendMutation(updateInvitationMutation)
		return nil
	})

	if err != nil {
		return entity.Invitation{}, err
	}

	return invitation, nil
}

func (i Invitation) DeleteInvitation(ct context.Context, invitationID uint64) (entity.Invitation, *errs.Error) {
	invitation, err := i.invitationDaoV2.FindInvitationByID(ct, invitationID)
	if err != nil {
		return entity.Invitation{}, err
	}

	txCtx := TransactionsContext{
		logger:             i.logger,
		transactionFactory: i.transactionFactory,
		stateSyncer:        i.stateSyncer,
		ct:                 ct,
	}
	err = txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		deleteInvitationMutation := mutation.NewDeleteInvitation(
			i.logger,
			i.stateSyncer,
			i.invitationDao,
			i.invitationDaoV2,
			invitation,
		)
		internalErr := deleteInvitationMutation.ExecuteV2(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		rtTx.AppendMutation(deleteInvitationMutation)
		return nil
	})

	if err != nil {
		return entity.Invitation{}, err
	}

	return invitation, nil
}

func (i Invitation) AcceptInvitation(ct context.Context, invitationID uint64, invitationCode string) (entity.Invitation, *errs.Error) {
	receiverUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Invitation{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	invitation, err := i.invitationDaoV2.FindInvitationByID(ct, invitationID)
	if err != nil {
		return entity.Invitation{}, err
	}

	if i.featureToggles.EnableAuthorization {
		query := authorization.NewAddMemberToInTeamQuery(receiverUserID, invitation.TeamID)
		hasPermission, internalErr := i.authorizer.hasPermission(ct, query)
		if internalErr != nil {
			return entity.Invitation{}, err
		}

		if !hasPermission {
			return entity.Invitation{}, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	if invitation.Code != invitationCode {
		return entity.Invitation{}, errs.NewError(
			errs.InvalidArgument,
			fmt.Sprintf("invalid invitation code: invitationID=%v, invitationCode=%v",
				invitationID,
				invitationCode,
			))
	}

	err = i.ensureInvitationPending(ct, invitation)
	if err != nil {
		return entity.Invitation{}, err
	}

	invitation.Status = entity.InvitationStatusAccepted
	invitation.ReceiverUserID = &receiverUserID
	now := time.Now().UTC()
	invitation.UpdatedAt = &now
	txCtx := TransactionsContext{
		logger:             i.logger,
		transactionFactory: i.transactionFactory,
		stateSyncer:        i.stateSyncer,
		ct:                 ct,
	}
	err = txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		updateInvitationMutation := mutation.NewUpdateInvitation(
			i.logger,
			i.stateSyncer,
			i.invitationDao,
			i.invitationDaoV2,
			invitation,
		)
		internalErr := updateInvitationMutation.ExecuteV2(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		rtTx.AppendMutation(updateInvitationMutation)
		_, err = i.teamMemberDaoV2.FindTeamMemberWithTx(ct, tx, invitation.TeamID, receiverUserID)
		if err != nil {
			if err.Code != errs.NotFound {
				return err
			}

			teamMember := entity.TeamMember{
				TeamID:    invitation.TeamID,
				UserID:    receiverUserID,
				CreatedAt: now,
			}
			createTeamMemberMutation := mutation.NewCreateTeamMember(
				i.logger,
				i.stateSyncer,
				i.teamMemberDao,
				i.teamMemberDaoV2,
				teamMember,
			)
			internalErr = createTeamMemberMutation.ExecuteV2(ct, tx)
			if internalErr != nil {
				return internalErr
			}

			rtTx.AppendMutation(createTeamMemberMutation)

			var sprints []entity.Sprint
			sprints, internalErr = i.sprintDaoV2.FindSprintsByTeamIDWithTx(ct, tx, invitation.TeamID)
			if internalErr != nil {
				return internalErr
			}

			now := time.Now().UTC()
			currAndFutureSprints := collect.Filter(sprints, func(sprint entity.Sprint) bool {
				if sprint.EndAt.UTC().Before(now) {
					return false
				}

				return true
			})

			for _, sprint := range currAndFutureSprints {
				participant := entity.SprintParticipant{
					SprintID:  sprint.ID,
					UserID:    receiverUserID,
					CreatedAt: now,
				}
				createSprintParticipantMutation := mutation.NewCreateSprintParticipant(
					i.logger,
					i.stateSyncer,
					i.sprintParticipantDao,
					i.sprintParticipantDaoV2,
					i.sprintDao,
					i.sprintDaoV2,
					participant,
				)
				internalErr = createTeamMemberMutation.ExecuteV2(ct, tx)
				if internalErr != nil {
					return internalErr
				}

				rtTx.AppendMutation(createSprintParticipantMutation)
			}
		}

		return nil
	})

	if err != nil {
		return entity.Invitation{}, err
	}

	return invitation, nil
}

func (i Invitation) DeclineInvitation(ct context.Context, invitationID uint64, invitationCode string) (entity.Invitation, *errs.Error) {
	receiverUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Invitation{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	invitation, err := i.invitationDaoV2.FindInvitationByID(ct, invitationID)
	if err != nil {
		return entity.Invitation{}, err
	}

	if invitation.Code != invitationCode {
		return entity.Invitation{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("invalid invitation code: invitationID=%v, invitationCode=%v",
			invitationID,
			invitationCode,
		))
	}

	err = i.ensureInvitationPending(ct, invitation)
	if err != nil {
		return entity.Invitation{}, err
	}

	invitation.Status = entity.InvitationStatusDeclined
	invitation.ReceiverUserID = &receiverUserID
	now := time.Now().UTC()
	invitation.UpdatedAt = &now
	txCtx := TransactionsContext{
		logger:             i.logger,
		transactionFactory: i.transactionFactory,
		stateSyncer:        i.stateSyncer,
		ct:                 ct,
	}
	err = txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		updateInvitationMutation := mutation.NewUpdateInvitation(
			i.logger,
			i.stateSyncer,
			i.invitationDao,
			i.invitationDaoV2,
			invitation,
		)

		err = updateInvitationMutation.ExecuteV2(ct, tx)
		if err != nil {
			return err
		}

		rtTx.AppendMutation(updateInvitationMutation)
		return nil
	})

	if err != nil {
		return entity.Invitation{}, err
	}

	return invitation, nil
}

func (i Invitation) ensureInvitationPending(ct context.Context, invitation entity.Invitation) *errs.Error {
	switch invitation.Status {
	case entity.InvitationStatusExpired:
		return errs.NewError(
			errs.InvalidOperation,
			fmt.Sprintf("invitation is expired: invitationID=%v", invitation.ID))
	case entity.InvitationStatusInvoked:
		return errs.NewError(
			errs.InvalidOperation,
			fmt.Sprintf("invitation is revoked: invitationID=%v", invitation.ID))
	case entity.InvitationStatusAccepted, entity.InvitationStatusDeclined:
		return errs.NewError(
			errs.InvalidOperation,
			fmt.Sprintf("invitation is already responded: invitationID=%v", invitation.ID))
	default:
		return nil
	}
}

func NewInvitation(
	logger telemetry.Logger,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	authorizer Authorizer,
	featureToggles feature.Toggles,
	stateSyncer *realtime.StateSyncer,
	transactionFactory transaction.Factory,
	invitationDao dao.Invitation,
	invitationDaoV2 daov2.Invitation,
	teamMemberDao dao.TeamMember,
	teamMemberDaoV2 daov2.TeamMember,
	sprintParticipantDao dao.SprintParticipant,
	sprintParticipantDaoV2 daov2.SprintParticipant,
	sprintDao dao.Sprint,
	sprintDaoV2 daov2.Sprint,
) Invitation {
	return Invitation{
		logger:                 logger,
		cloudClientRegistry:    cloudClientRegistry,
		authorizer:             authorizer,
		featureToggles:         featureToggles,
		stateSyncer:            stateSyncer,
		transactionFactory:     transactionFactory,
		invitationDao:          invitationDao,
		invitationDaoV2:        invitationDaoV2,
		teamMemberDao:          teamMemberDao,
		teamMemberDaoV2:        teamMemberDaoV2,
		sprintParticipantDao:   sprintParticipantDao,
		sprintParticipantDaoV2: sprintParticipantDaoV2,
		sprintDao:              sprintDao,
		sprintDaoV2:            sprintDaoV2,
	}
}
