package gql

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/api/gql/scalar"
	"github.com/teamyapp/teamy-backend/core/entity"
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

	teamMember := entity.TeamMember{
		TeamID:    teamID,
		UserID:    memberUserID,
		CreatedAt: time.Now(),
	}
	err = m.deps.teamMemberSyncer.CreateAndSyncTeamMember(ct, teamMember)
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

	teamMember, err := m.deps.teamMemberDao.FindTeamMember(ct, teamID, memberUserID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return TeamMember{}, err
	}

	teamMember.WeeklyBandwidth = args.Input.WeeklyBandwidth.Duration
	now := time.Now()
	teamMember.UpdatedAt = &now

	err = m.deps.teamMemberDao.UpdateTeamMember(ct, teamMember)
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

	teamMember, err := m.deps.teamMemberDao.FindTeamMember(ct, teamID, memberUserID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return TeamMember{}, err
	}

	// TODO: ensure user is inside the team
	err = m.deps.teamMemberSyncer.DeleteAndSyncTeamMember(ct, teamID, memberUserID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return TeamMember{}, err
	}

	return newTeamMember(m.deps, teamMember), nil
}
