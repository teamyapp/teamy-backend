package service

import (
	"context"
	"errors"
	"fmt"
	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/cloud/libs/randgen"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/collection"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"time"
)

const invitationCodeLen = 20

type Invitation struct {
	dataCollector       obs.DataCollector
	cloudClientRegistry *cloudAPI.ClientRegistry
	authorizer          Authorizer
	invitationSyncer    collection.InvitationSyncer
}

type CreateInvitationInput struct {
	ReceiverFirstName *string
	ReceiverLastName  *string
	ReceiverEmail     *string
	ExpireAt          time.Time
}

func (i Invitation) CreateInvitation(ct context.Context, teamID uint64, input CreateInvitationInput) (entity.Invitation, error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		i.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Invitation{}, err
	}

	if feature.EnableAuthorization {
		query := authorization.NewCreateInvitationQuery(userID, teamID)
		hasPermission, err := i.authorizer.hasPermission(ct, query)
		if err != nil {
			i.authorizer.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Invitation{}, err
		}

		if !hasPermission {
			return entity.Invitation{}, authorization.Error{
				Code:    authorization.UnauthorizedErrorCode,
				Message: fmt.Sprintf("Unauthorized: %v", query),
			}
		}
	}

	genInvitationIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "invitationID"}
	genInvitationIDRes, err := i.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genInvitationIDReq)
	if err != nil {
		i.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Invitation{}, err
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

	err = i.invitationSyncer.CreateAndSyncInvitation(ct, invitation)
	if err != nil {
		i.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Invitation{}, err
	}

	if feature.EnableAuthorization {
		err = i.authorizer.registerResource(ct, authorization.InvitationResourceType, invitation.ID)
		if err != nil {
			i.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Invitation{}, err
		}

		err = i.authorizer.assignParentResource(ct, authorization.InvitationResourceType, invitation.ID, authorization.TeamResourceType, invitation.TeamID)
		if err != nil {
			i.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Invitation{}, err
		}
	}

	return invitation, nil
}

func NewInvitation(
	dataCollector obs.DataCollector,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	authorizer Authorizer,
	invitationSyncer collection.InvitationSyncer,
) Invitation {
	return Invitation{
		dataCollector:       dataCollector,
		cloudClientRegistry: cloudClientRegistry,
		authorizer:          authorizer,
		invitationSyncer:    invitationSyncer,
	}
}
