package resolver

import (
	"fmt"
	"log"

	"github.com/teamyapp/teamy-backend/app/entity"
)

type Team struct {
	Entity
	deps *Dependencies
	team entity.Team
}

func (t Team) Name() string {
	return t.team.Name
}

func (t Team) LogoURL() *string {
	return t.team.LogoURL
}

func (t Team) Members() ([]User, error) {
	ids, err := t.deps.teamRepo.ListTeamMemberIDs(t.team.ID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	members, err := t.deps.userRepo.FindUsers(ids)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return toGraphQLUsers(t.deps, members), nil
}

func (t Team) Tasks(args struct{ Input *TaskFilter }) ([]Task, error) {
	teams := t.deps.Data.FilterTeams(func(team entity.Team) bool {
		return team.ID == t.entity.ID
	})
	if len(teams) != 1 {
		return nil, fmt.Errorf("more than 1 team found for %v", t.entity.ID)
	}
	team := teams[0]
	tasks := t.deps.Data.FilterTasks(func(t entity.Task) bool {
		for _, taskID := range team.Tasks {
			if t.ID == taskID {
				return true
			}
		}
		return false
	})
	return newTasks(t.deps, tasks), nil
}

func newTeam(deps *Dependencies, team entity.Team) Team {
	return Team{
		Entity: Entity{entity: team.Entity},
		deps:   deps,
		team:   team,
	}
}
