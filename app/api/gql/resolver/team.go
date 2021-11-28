package resolver

import (
	"github.com/teamyapp/teamy-backend/app/entity"
	"github.com/teamyapp/teamy-backend/app/service"
	"log"
)

type Team struct {
	Entity
	team entity.Team
	userService service.User
}

func (t Team) Members() ([]User, error) {
	members, err := t.userService.ListTeamMembers(t.team.ID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return toGraphQLUsers(members), nil
}

func newTeam(team entity.Team, userService service.User) Team {
	return Team{
		Entity: Entity{entity: team.Entity},
		team: team,
		userService: userService,
	}
}

