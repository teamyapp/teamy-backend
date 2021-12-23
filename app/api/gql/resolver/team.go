package resolver

import (
	"context"
	"fmt"
	"sort"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/one/identity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Team struct {
	deps *Dependencies
	entity.Team
}

func (t Team) CreatedAt() graphql.Time {
	return graphql.Time{Time: t.Team.CreatedAt}
}

func (t Team) ID() graphql.ID {
	return toGraphQLID(t.Team.ID)
}

func (t Team) Name() string {
	return t.Team.Name
}

func (t Team) IconURL() string {
	return t.Team.IconURL
}

func (t Team) Members() ([]User, error) {
	members, err := t.deps.Data.GetUsers(t.Team.MemberIDs)
	return toGraphQLUsers(t.deps, members), err
}

func (t Team) Tasks(args struct{ Input *TaskFilter }) ([]Task, error) {
	tasks := t.deps.Data.FilterTasks(func(task entity.Task) bool {
		for _, taskID := range t.Team.Tasks {
			if task.ID == taskID {
				return taskFilterFunc(task, args.Input)
			}
		}
		return false
	})
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
	return newTasks(t.deps, tasks), nil
}

func (t Team) Creator() (User, error) {
	fmt.Println(t.Team.CreatorID)
	user, err := t.deps.Data.GetUser(t.Team.CreatorID)
	return newUser(t.deps, user), err
}
func (t Team) Owner() (User, error) {
	user, err := t.deps.Data.GetUser(t.Team.CreatorID)
	return newUser(t.deps, user), err
}
func (t Team) Admins() ([]User, error) {
	user, err := t.deps.Data.GetUser(t.Team.CreatorID)
	return []User{newUser(t.deps, user)}, err
}

func (t Team) TasksNeedAttention(ctx context.Context, args struct{ IsMine bool }) ([]Task, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	tasks := t.deps.Data.FilterTasks(func(task entity.Task) bool {
		for _, taskID := range t.NeedAttentionTasks {
			if taskID == task.ID && task.Status == entity.IN_PROGRESS {
				if args.IsMine {
					if task.OwnerUserId == nil {
						return false
					} else {
						return userID == *task.OwnerUserId
					}
				} else {
					return true
				}
			}
		}
		return false
	})
	return newTasks(t.deps, tasks), nil
}

func newTeam(deps *Dependencies, team entity.Team) Team {
	return Team{
		deps: deps,
		Team: team,
	}
}

func newTeams(deps *Dependencies, teams []entity.Team) (ts []Team) {
	for _, t := range teams {
		ts = append(ts, newTeam(deps, t))
	}
	return
}
