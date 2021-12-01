package resolver

import (
	"log"

	"github.com/teamyapp/teamy-backend/app/api/gqlv2/resolver"

	"github.com/teamyapp/teamy-backend/app/entity"
)

type Team struct {
	Entity
	deps          *Dependencies
	prototypeDeps *resolver.Dependencies
	team          entity.Team
}

func (t Team) Members() ([]User, error) {
	members, err := t.deps.teamService.ListTeamMembers(t.team.ID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return toGraphQLUsers(members), nil
}

func newTeam(deps *Dependencies, prototypeDeps *resolver.Dependencies, team entity.Team) Team {
	return Team{
		Entity:        Entity{entity: team.Entity},
		deps:          deps,
		prototypeDeps: prototypeDeps,
		team:          team,
	}
}
