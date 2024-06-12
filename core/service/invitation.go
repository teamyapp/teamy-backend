package service

import (
	"context"
	"fmt"
	"time"

	"github.com/teamyapp/cloud/app/client"
	cloudAuthorization "github.com/teamyapp/cloud/libs/authorization"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/randgen"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudTransaction "github.com/teamyapp/cloud/libs/transaction"
	pbcloud "github.com/teamyapp/protocol/pb/pbgo/cloud"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/transaction"
)

const invitationCodeLen = 20

type Invitation struct {
	logger                               telemetry.Logger
	transactionGroupFactory              transaction.GroupFactory
	cloudClientRegistry                  *client.Registry
	authorizer                           client.Authorizer
	featureToggles                       feature.Toggles
	stateSyncer                          *realtime.StateSyncer
	transactionFactory                   cloudTransaction.Factory
	invitationDao                        dao.Invitation
	teamMemberGroupInvitationRelationDao dao.TeamMemberGroupInvitationRelation
	teamMemberDao                        dao.TeamMember
	teamMemberGroupDao                   dao.TeamMemberGroup
	sprintParticipantDao                 dao.SprintParticipant
	sprintDao                            dao.Sprint
	teamDao                              dao.Team
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
	userID, ok := ctx.UserIDFromContext(ct)
	if i.featureToggles.EnableAuthorization {
		if !ok {
			return nil, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewReadInTeamQuery(userID, teamID)
		hasPermission, err := i.authorizer.HasPermission(ct, query)
		if err != nil {
			return nil, err
		}

		if !hasPermission {
			return nil, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	invitations, err := i.invitationDao.FindInvitationsByTeamID(ct, teamID)
	if err != nil {
		return nil, err
	}

	if i.featureToggles.EnableAuthorization {
		authorizedInvitations, err := client.FilterAuthorizedItems(
			ct,
			i.authorizer,
			invitations,
			func(invitation entity.Invitation) cloudAuthorization.Query {
				return authorization.NewReadInInvitationQuery(userID, invitation.ID)
			})
		if err != nil {
			return nil, err
		}

		invitations = authorizedInvitations
	}

	if filter != nil {
		invitations = filterInvitations(invitations, *filter)
	}

	return invitations, nil
}

func (i Invitation) FindInvitations(ct context.Context, filter *InvitationFilter) ([]entity.Invitation, *errs.Error) {
	invitations, err := i.invitationDao.FindAllInvitations(ct)
	if err != nil {
		return nil, err
	}

	if i.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return nil, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		authorizedInvitations, err := client.FilterAuthorizedItems(
			ct,
			i.authorizer,
			invitations,
			func(invitation entity.Invitation) cloudAuthorization.Query {
				return authorization.NewReadInInvitationQuery(userID, invitation.ID)
			})
		if err != nil {
			return nil, err
		}

		invitations = authorizedInvitations
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
		hasPermission, err := i.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Invitation{}, err
		}

		if !hasPermission {
			return entity.Invitation{}, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	genInvitationIDReq := &pbcloud.GenerateUniqueNumberRequest{SequenceName: "invitationID"}
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

	err := i.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		createInvitationMutation := mutation.NewCreateInvitation(
			i.logger,
			i.stateSyncer,
			i.invitationDao,
			invitation,
		)

		internalErr := createInvitationMutation.Execute(ct, tx)
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
		err = i.authorizer.RegisterResource(ct, authorization.InvitationResourceType, invitation.ID)
		if err != nil {
			return entity.Invitation{}, err
		}

		err = i.authorizer.AssignParentResource(ct, authorization.InvitationResourceType, invitation.ID, authorization.TeamResourceType, invitation.TeamID)
		if err != nil {
			return entity.Invitation{}, err
		}
	}

	return invitation, nil
}

func (i Invitation) UpdateInvitation(ct context.Context, invitationID uint64, input UpdateInvitationInput) (entity.Invitation, *errs.Error) {
	if i.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.Invitation{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewUpdateInInvitationQuery(userID, invitationID)
		hasPermission, err := i.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Invitation{}, err
		}

		if !hasPermission {
			return entity.Invitation{}, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var invitation entity.Invitation
	err := i.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		invitation, internalErr = i.invitationDao.FindInvitationByIDWithTx(ct, tx, invitationID)
		if internalErr != nil {
			return internalErr
		}

		invitation.ReceiverFirstName = input.ReceiverFirstName
		invitation.ReceiverLastName = input.ReceiverLastName
		invitation.ExpireAt = input.ExpireAt
		now := time.Now().UTC()
		invitation.UpdatedAt = &now

		updateInvitationMutation := mutation.NewUpdateInvitation(
			i.logger,
			i.stateSyncer,
			i.invitationDao,
			invitation,
		)
		internalErr = updateInvitationMutation.Execute(ct, tx)
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
	if i.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.Invitation{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewDeleteInInvitationQuery(userID, invitationID)
		hasPermission, err := i.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Invitation{}, err
		}

		if !hasPermission {
			return entity.Invitation{}, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var invitation entity.Invitation
	err := i.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		invitation, internalErr = i.invitationDao.FindInvitationByID(ct, invitationID)
		if internalErr != nil {
			return internalErr
		}

		deleteInvitationMutation := mutation.NewDeleteInvitation(
			i.logger,
			i.stateSyncer,
			i.invitationDao,
			invitation,
		)
		internalErr = deleteInvitationMutation.Execute(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		rtTx.AppendMutation(deleteInvitationMutation)
		return nil
	})

	if err != nil {
		return entity.Invitation{}, err
	}

	// TODO: update resource relations in authorization service
	return invitation, nil
}

func (i Invitation) AcceptInvitation(ct context.Context, invitationID uint64, invitationCode string) (entity.Invitation, *errs.Error) {
	invitation, err := i.canRespondToInvitation(ct, invitationID, invitationCode)
	if err != nil {
		return entity.Invitation{}, err
	}

	receiverUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Invitation{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	invitation.Status = entity.InvitationStatusAccepted
	invitation.ReceiverUserID = &receiverUserID
	now := time.Now().UTC()
	invitation.UpdatedAt = &now
	err = i.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		updateInvitationMutation := mutation.NewUpdateInvitation(
			i.logger,
			i.stateSyncer,
			i.invitationDao,
			invitation,
		)
		internalErr := updateInvitationMutation.Execute(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		rtTx.AppendMutation(updateInvitationMutation)
		_, err = i.teamMemberDao.FindTeamMemberWithTx(ct, tx, invitation.TeamID, receiverUserID)
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
				teamMember,
			)
			internalErr = createTeamMemberMutation.Execute(ct, tx)
			if internalErr != nil {
				return internalErr
			}

			rtTx.AppendMutation(createTeamMemberMutation)

			var sprints []entity.Sprint
			sprints, internalErr = i.sprintDao.FindSprintsByTeamIDWithTx(ct, tx, invitation.TeamID)
			if internalErr != nil {
				return internalErr
			}

			now = time.Now().UTC()
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
					i.sprintDao,
					participant,
				)
				internalErr = createSprintParticipantMutation.Execute(ct, tx)
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
	invitation, err := i.canRespondToInvitation(ct, invitationID, invitationCode)
	if err != nil {
		return entity.Invitation{}, err
	}

	receiverUserID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Invitation{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	invitation.Status = entity.InvitationStatusDeclined
	invitation.ReceiverUserID = &receiverUserID
	now := time.Now().UTC()
	invitation.UpdatedAt = &now
	err = i.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		updateInvitationMutation := mutation.NewUpdateInvitation(
			i.logger,
			i.stateSyncer,
			i.invitationDao,
			invitation,
		)

		err = updateInvitationMutation.Execute(ct, tx)
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

func (i Invitation) FindInvitationsByTeamMemberGroupID(ct context.Context, teamMemberGroupID uint64) ([]entity.Invitation, *errs.Error) {
	var invitations []entity.Invitation
	err := i.transactionGroupFactory.WithTransactionGroup(ct, true, func(tx *cloudTransaction.Transaction, _ *realtime.Transaction) *errs.Error {
		invitationIDs, internalErr := i.teamMemberGroupInvitationRelationDao.FindInvitationIDsByTeamMemberGroupID(ct, tx, teamMemberGroupID)
		if internalErr != nil {
			return internalErr
		}

		invitations, internalErr = i.invitationDao.FindInvitationsByIDsWithTx(ct, tx, invitationIDs)
		return internalErr
	})

	return invitations, err
}

func (i Invitation) AddInvitationToTeamMemberGroup(ct context.Context, invitationID uint64, teamMemberGroupID uint64) (entity.Invitation, *errs.Error) {
	var invitation entity.Invitation
	err := i.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		invitation, internalErr = i.invitationDao.FindInvitationByID(ct, invitationID)
		if internalErr != nil {
			return internalErr
		}

		_, internalErr = i.teamMemberGroupDao.FindMemberGroupByID(ct, tx, teamMemberGroupID)
		if internalErr != nil {
			return internalErr
		}

		teamMemberGroupInvitationRelation := entity.TeamMemberGroupInvitationRelation{
			GroupID:      teamMemberGroupID,
			InvitationID: invitationID,
			CreatedAt:    time.Now().UTC(),
		}
		createTeamMemberGroupInvitationRelationMutation := mutation.NewCreateTeamMemberGroupInvitationRelation(
			i.logger,
			i.stateSyncer,
			i.teamMemberGroupInvitationRelationDao,
			i.teamMemberGroupDao,
			i.teamDao,
			teamMemberGroupInvitationRelation,
		)

		internalErr = createTeamMemberGroupInvitationRelationMutation.Execute(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		rtTx.AppendMutation(createTeamMemberGroupInvitationRelationMutation)
		return nil
	})

	return invitation, err
}

func (i Invitation) RemoveInvitationFromTeamMemberGroup(ct context.Context, invitationID uint64, teamMemberGroupID uint64) (entity.Invitation, *errs.Error) {
	var invitation entity.Invitation
	err := i.transactionGroupFactory.WithTransactionGroup(ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		invitation, internalErr = i.invitationDao.FindInvitationByID(ct, invitationID)
		if internalErr != nil {
			return internalErr
		}

		teamMemberGroupInvitationRelation := entity.TeamMemberGroupInvitationRelation{
			GroupID:      teamMemberGroupID,
			InvitationID: invitationID,
		}

		return i.teamMemberGroupInvitationRelationDao.DeleteTeamMemberGroupInvitationRelation(ct, tx, teamMemberGroupInvitationRelation)
	})

	return invitation, err
}

func (i Invitation) canRespondToInvitation(
	ct context.Context,
	invitationID uint64,
	invitationCode string,
) (entity.Invitation, *errs.Error) {
	invitation, err := i.invitationDao.FindInvitationByID(ct, invitationID)
	if err != nil {
		return entity.Invitation{}, err
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
	transactionGroupFactory transaction.GroupFactory,
	cloudClientRegistry *client.Registry,
	authorizer client.Authorizer,
	featureToggles feature.Toggles,
	stateSyncer *realtime.StateSyncer,
	transactionFactory cloudTransaction.Factory,
	invitationDao dao.Invitation,
	teamMemberGroupInvitationRelationDao dao.TeamMemberGroupInvitationRelation,
	teamMemberDao dao.TeamMember,
	teamMemberGroupDao dao.TeamMemberGroup,
	sprintParticipantDao dao.SprintParticipant,
	sprintDao dao.Sprint,
	teamDao dao.Team,
) Invitation {
	return Invitation{
		logger:                               logger,
		transactionGroupFactory:              transactionGroupFactory,
		cloudClientRegistry:                  cloudClientRegistry,
		authorizer:                           authorizer,
		featureToggles:                       featureToggles,
		stateSyncer:                          stateSyncer,
		transactionFactory:                   transactionFactory,
		invitationDao:                        invitationDao,
		teamMemberGroupInvitationRelationDao: teamMemberGroupInvitationRelationDao,
		teamMemberDao:                        teamMemberDao,
		teamMemberGroupDao:                   teamMemberGroupDao,
		sprintParticipantDao:                 sprintParticipantDao,
		sprintDao:                            sprintDao,
		teamDao:                              teamDao,
	}
}
