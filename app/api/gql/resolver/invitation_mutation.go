package resolver

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/one/identity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type InvitationUpdate struct {
	deps       *Dependencies
	invitation entity.Invitation
}

type InvitationInput struct {
	SenderUserID   *graphql.ID
	ReceiverUserID *graphql.ID
	TeamID         *graphql.ID
	ExpireAt       *graphql.Time
	Status         *entity.InvitationStatus
	Code           *string
}

func (i InvitationUpdate) Invitation(ctx context.Context) Invitation {
	return Invitation{deps: i.deps, Invitation: i.invitation}
}

func (i InvitationUpdate) Accept(ctx context.Context) (InvitationUpdate, error) {
	switch i.invitation.Status {
	case entity.InvitationStatusInvoked:
		return InvitationUpdate{}, fmt.Errorf("invitation is revoked: %v", i.invitation.ID)
	case entity.InvitationStatusExpired:
		return InvitationUpdate{}, fmt.Errorf("invitation is expired: %v", i.invitation.ID)
	case entity.InvitationStatusAccepted:
		return InvitationUpdate{}, fmt.Errorf("invitation has been accepted already: %v", i.invitation.ID)
	case entity.InvitationStatusDeclined:
		return InvitationUpdate{}, fmt.Errorf("invitation has been declined already: %v", i.invitation.ID)
	}

	if time.Now().After(i.invitation.ExpireAt) {
		updated, err := i.deps.Data.UpdateInvitation(i.invitation.ID, func(invitation entity.Invitation) entity.Invitation {
			invitation.Status = entity.InvitationStatusExpired
			return invitation
		})
		if err != nil {
			return InvitationUpdate{}, err
		}
		return InvitationUpdate{}, fmt.Errorf("invitation expired: %v", updated.ID)
	}

	userID, err := identity.FromContext(ctx)
	if err != nil {
		return InvitationUpdate{}, err
	}

	teams := i.deps.Data.FilterTeams(func(team entity.Team) bool {
		return team.ID == i.invitation.TeamID
	})
	if len(teams) != 1 {
		return InvitationUpdate{}, fmt.Errorf("should find only 1 team: %v", i.invitation.TeamID)
	}

	// TODO: use transaction to update invitation & team together
	invitation, err := i.deps.Data.UpdateInvitation(
		i.invitation.ID,
		func(invitation entity.Invitation) entity.Invitation {
			invitation.ReceiverUserID = &userID
			invitation.Status = entity.InvitationStatusAccepted
			return invitation
		})
	if err != nil {
		return InvitationUpdate{}, err
	}

	tu := TeamUpdate{team: teams[0], deps: i.deps}
	_, err = tu.AddMember(struct{ ID graphql.ID }{ID: toGraphQLID(userID)})
	if err != nil {
		return InvitationUpdate{}, err
	}

	i.invitation = invitation
	return i, nil
}

func (i InvitationUpdate) Decline(ctx context.Context) (InvitationUpdate, error) {
	switch i.invitation.Status {
	case entity.InvitationStatusInvoked:
		return InvitationUpdate{}, fmt.Errorf("invitation is revoked: %v", i.invitation.ID)
	case entity.InvitationStatusExpired:
		return InvitationUpdate{}, fmt.Errorf("invitation is expired: %v", i.invitation.ID)
	case entity.InvitationStatusAccepted:
		return InvitationUpdate{}, fmt.Errorf("invitation has been accepted already: %v", i.invitation.ID)
	case entity.InvitationStatusDeclined:
		return InvitationUpdate{}, fmt.Errorf("invitation has been declined already: %v", i.invitation.ID)
	}

	userID, err := identity.FromContext(ctx)
	if err != nil {
		return InvitationUpdate{}, err
	}

	invitation, err := i.deps.Data.UpdateInvitation(
		i.invitation.ID,
		func(invitation entity.Invitation) entity.Invitation {
			invitation.ReceiverUserID = &userID
			invitation.Status = entity.InvitationStatusDeclined
			return invitation
		})
	if err != nil {
		return InvitationUpdate{}, err
	}
	i.invitation = invitation
	return i, nil
}

func (i InvitationUpdate) Update(ctx context.Context, args struct {
	Input InvitationInput
}) (InvitationUpdate, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		return InvitationUpdate{}, err
	}
	teams := i.deps.Data.FilterTeams(func(team entity.Team) bool {
		return team.ID == i.invitation.TeamID && team.CreatorID == userID
	})
	if len(teams) != 1 {
		return InvitationUpdate{}, fmt.Errorf("should find only 1 team: %v", i.invitation.TeamID)
	}

	updated, err := i.deps.Data.UpdateInvitation(i.invitation.ID, func(invitation entity.Invitation) entity.Invitation {
		if args.Input.SenderUserID != nil {
			invitation.SenderUserID, err = fromGraphQLID(*args.Input.SenderUserID)
			if err != nil {
				return invitation
			}
		}

		if args.Input.ReceiverUserID != nil {
			invitation.ReceiverUserID, err = fromGraphQLIDPtr(args.Input.ReceiverUserID)
			if err != nil {
				return invitation
			}
		}

		if args.Input.TeamID != nil {
			invitation.TeamID, err = fromGraphQLID(*args.Input.TeamID)
			if err != nil {
				return invitation
			}
		}

		if args.Input.ExpireAt != nil {
			invitation.ExpireAt = (*args.Input.ExpireAt).Time
		}

		if args.Input.Status != nil {
			invitation.Status = *args.Input.Status
		}

		if args.Input.Code != nil {
			invitation.Code = *args.Input.Code
		}
		return invitation
	})
	if err != nil {
		return InvitationUpdate{}, err
	}
	i.invitation = updated
	return i, nil
}

func (i InvitationUpdate) Delete(ctx context.Context) (Invitation, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		return Invitation{}, err
	}
	teams := i.deps.Data.FilterTeams(func(team entity.Team) bool {
		return team.ID == i.invitation.TeamID && team.CreatorID == userID
	})
	if len(teams) != 1 {
		return Invitation{}, fmt.Errorf("should find only 1 team: %v", i.invitation.TeamID)
	}

	deleted, err := i.deps.Data.DeleteInvitation(i.invitation.ID)
	return Invitation{deps: i.deps, Invitation: deleted}, err
}

func (m Mutation) CreateInvitation(ctx context.Context, args struct {
	Input struct {
		ReceiverFirstName string
		ReceiverLastName  string
		TeamID            graphql.ID
		ExpireAt          graphql.Time
	}
}) (InvitationUpdate, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		return InvitationUpdate{}, err
	}

	teamID, err := fromGraphQLID(args.Input.TeamID)
	if err != nil {
		return InvitationUpdate{}, err
	}

	teams := m.deps.Data.FilterTeams(func(team entity.Team) bool {
		return team.ID == teamID && team.CreatorID == userID
	})
	if len(teams) != 1 {
		return InvitationUpdate{}, fmt.Errorf("should find only 1 team: %v", teamID)
	}

	if err != nil {
		return InvitationUpdate{}, err
	}

	invitation := entity.Invitation{
		SenderUserID:      userID,
		ReceiverFirstName: args.Input.ReceiverFirstName,
		ReceiverLastName:  args.Input.ReceiverLastName,
		TeamID:            teamID,
		Status:            entity.InvitationStatusPending,
		Code:              uuid.New().String(),
		ExpireAt:          args.Input.ExpireAt.Time,
	}

	created, err := m.deps.Data.CreateInvitation(invitation)
	if err != nil {
		return InvitationUpdate{}, err
	}

	_, err = m.deps.Data.UpdateTeam(teamID, func(team entity.Team) entity.Team {
		team.InvitationIDs = team.InvitationIDs.Add(created.ID)
		return team
	})
	if err != nil {
		return InvitationUpdate{}, err
	}

	return InvitationUpdate{
		deps:       m.deps,
		invitation: created,
	}, nil
}

func (m Mutation) InvitationUpdate(ctx context.Context, args struct {
	ID graphql.ID
}) (InvitationUpdate, error) {
	id, err := fromGraphQLID(args.ID)
	if err != nil {
		return InvitationUpdate{}, err
	}

	invitations := m.deps.Data.FilterInvitations(func(invitation entity.Invitation) bool {
		return invitation.ID == id
	})
	if len(invitations) != 1 {
		return InvitationUpdate{}, fmt.Errorf("should find only 1 Invitation: %v", id)
	}
	return InvitationUpdate{deps: m.deps, invitation: invitations[0]}, nil
}

func (m Mutation) InvitationUpdateWithCode(ctx context.Context, args struct {
	Code string
}) (InvitationUpdate, error) {
	invitations := m.deps.Data.FilterInvitations(func(invitation entity.Invitation) bool {
		return invitation.Code == args.Code
	})
	if len(invitations) != 1 {
		return InvitationUpdate{}, fmt.Errorf("should find only 1 Invitation: %v", args.Code)
	}
	return InvitationUpdate{deps: m.deps, invitation: invitations[0]}, nil
}
