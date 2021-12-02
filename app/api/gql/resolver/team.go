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
	ids, err := t.deps.teamRepo.ListTeamMemberIDs(t.team.ID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	members, err := t.deps.userRepo.GetUsers(ids)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return toGraphQLUsers(t.deps, t.prototypeDeps, members), nil
}

func newTeam(deps *Dependencies, prototypeDeps *resolver.Dependencies, team entity.Team) Team {
	return Team{
		Entity:        Entity{entity: team.Entity},
		deps:          deps,
		prototypeDeps: prototypeDeps,
		team:          team,
	}
}
