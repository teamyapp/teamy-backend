package resolver

import (
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Invitation struct {
	deps *Dependencies
	entity.Invitation
}

func (i Invitation) ID() graphql.ID {
	return toGraphQLID(i.Invitation.ID)
}

func (i Invitation) Sender() (User, error) {
	user, err := i.deps.Data.GetUser(i.Invitation.SenderUserID)
	if err != nil {
		return User{}, err
	}

	return User{
		deps: i.deps,
		user: user,
	}, nil
}

func (i Invitation) Receiver() (*User, error) {
	if i.Invitation.ReceiverUserID == nil {
		return nil, nil
	}

	user, err := i.deps.Data.GetUser(*i.Invitation.ReceiverUserID)
	if err != nil {
		return &User{}, err
	}

	return &User{
		deps: i.deps,
		user: user,
	}, nil
}

func (i Invitation) Team() (Team, error) {
	teams := i.deps.Data.FilterTeams(func(team entity.Team) bool {
		return team.ID == i.Invitation.TeamID
	})
	if len(teams) < 1 {
		return Team{}, fmt.Errorf("team not found: teamID=%v", i.Invitation.TeamID)
	}
	if len(teams) > 1 {
		return Team{}, fmt.Errorf("more than 1 team found: teamID=%v", i.Invitation.TeamID)
	}

	return Team{deps: i.deps, Team: teams[0]}, nil
}

func (i Invitation) ExpireAt() graphql.Time {
	return graphql.Time{Time: i.Invitation.ExpireAt}
}

func (i Invitation) CreatedAt() graphql.Time {
	return graphql.Time{Time: i.Invitation.CreatedAt}
}
