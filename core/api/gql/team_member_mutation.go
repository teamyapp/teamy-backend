package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/api/gql/scalar"
	"github.com/teamyapp/teamy-backend/core/service"
)

func (m Mutation) AddMemberToTeam(ct context.Context, args struct {
	TeamID       graphql.ID
	MemberUserID graphql.ID
}) (TeamMember, error) {
	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return TeamMember{}, err
	}

	memberUserID, err := fromGraphQLID(args.MemberUserID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return TeamMember{}, err
	}

	teamMember, err := m.deps.teamService.AddMemberToTeam(ct, teamID, memberUserID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return TeamMember{}, err
	}

	return newTeamMember(m.deps, teamMember), nil
}

func (m Mutation) UpdateTeamMember(ct context.Context, args struct {
	TeamID graphql.ID
	Input  struct {
		UserID          graphql.ID
		WeeklyBandwidth scalar.Duration
	}
}) (TeamMember, error) {
	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return TeamMember{}, err
	}

	memberUserID, err := fromGraphQLID(args.Input.UserID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return TeamMember{}, err
	}

	input := service.UpdateTeamMemberInput{
		UserID:          memberUserID,
		WeeklyBandwidth: args.Input.WeeklyBandwidth.Duration,
	}
	teamMember, err := m.deps.teamService.UpdateTeamMember(ct, teamID, input)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return TeamMember{}, err
	}

	return newTeamMember(m.deps, teamMember), nil
}

func (m Mutation) RemoveMemberFromTeam(ct context.Context, args struct {
	TeamID       graphql.ID
	MemberUserID graphql.ID
}) (TeamMember, error) {
	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return TeamMember{}, err
	}

	memberUserID, err := fromGraphQLID(args.MemberUserID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return TeamMember{}, err
	}

	teamMember, err := m.deps.teamService.RemoveMemberFromTeam(ct, teamID, memberUserID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return TeamMember{}, err
	}

	return newTeamMember(m.deps, teamMember), nil
}
