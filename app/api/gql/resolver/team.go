package resolver

import (
	"fmt"

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

func (t Team) IconURL() *string {
	return t.team.IconURL
}

func (t Team) Members() ([]User, error) {
	members, err := t.deps.Data.GetUsers(t.team.MemberIDs)
	return toGraphQLUsers(t.deps, members), err
}

func (team Team) Tasks(args struct{ Input *TaskFilter }) ([]Task, error) {
	tasks := team.deps.Data.FilterTasks(func(t entity.Task) bool {
		for _, taskID := range team.team.Tasks {
			if t.ID == taskID {
				return taskFilterFunc(t, args.Input)
			}
		}
		return false
	})
	return newTasks(team.deps, tasks), nil
}

func (t Team) Creator() (User, error) {
	fmt.Println(t.team.CreatorID)
	user, err := t.deps.Data.GetUser(t.team.CreatorID)
	return newUser(t.deps, user), err
}
func (t Team) Owner() (User, error) {
	user, err := t.deps.Data.GetUser(t.team.CreatorID)
	return newUser(t.deps, user), err
}
func (t Team) Admins() ([]User, error) {
	user, err := t.deps.Data.GetUser(t.team.CreatorID)
	return []User{newUser(t.deps, user)}, err
}

func newTeam(deps *Dependencies, team entity.Team) Team {
	return Team{
		Entity: Entity{entity: team.Entity},
		deps:   deps,
		team:   team,
	}
}

func newTeams(deps *Dependencies, teams []entity.Team) (ts []Team) {
	for _, t := range teams {
		ts = append(ts, newTeam(deps, t))
	}
	return
}
