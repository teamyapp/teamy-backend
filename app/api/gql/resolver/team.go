package resolver

import (
	"context"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	oneEntity "github.com/teamyapp/one/entity"
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

func (t Team) IconURL() *string {
	return t.Team.IconURL
}

func (t Team) Members() ([]User, error) {
	members, err := t.deps.Data.GetUsers(t.Team.MemberIDs)
	return toGraphQLUsers(t.deps, members), err
}

func (team Team) Tasks(args struct{ Input *TaskFilter }) ([]Task, error) {
	tasks := team.deps.Data.FilterTasks(func(t entity.Task) bool {
		for _, taskID := range team.Team.Tasks {
			if t.ID == taskID {
				return taskFilterFunc(t, args.Input)
			}
		}
		return false
	})
	return newTasks(team.deps, tasks), nil
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

//////////////
// Mutation //
//////////////
func (m Mutation) CreateTeam(ctx context.Context,
	args struct {
		Input struct {
			Name    string
			IconURL *string
		}
	},
) (TeamUpdate, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		return TeamUpdate{}, err
	}
	t, err := m.deps.Data.CreateTeam(userID, entity.Team{
		Name:      args.Input.Name,
		IconURL:   args.Input.IconURL,
		CreatorID: userID,
		MemberIDs: []oneEntity.ID{
			userID,
		},
	})
	if err != nil {
		return TeamUpdate{}, err
	}
	return TeamUpdate{deps: m.deps, team: t}, nil
}

func (m Mutation) PromoteTaskToNeedAttention(
	ctx context.Context,
	args struct {
		TaskID graphql.ID
	},
) (Team, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		return Team{}, err
	}
	user, err := m.deps.Data.GetUser(userID)
	if err != nil {
		return Team{}, err
	}
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		return Team{}, err
	}
	task, err := m.deps.Data.GetTask(args.TaskID)
	if err != nil {
		return Team{}, err
	}
	if task.Status != entity.IN_PROGRESS {
		return Team{}, errors.Errorf("task %v is not in progress, can not promote to Need Attention", taskID)
	}
	team, err := m.deps.Data.UpdateTeam(user.ActiveTeamID, func(t entity.Team) entity.Team {
		t.NeedAttentionTasks[userID] = taskID
		return t
	})
	if err != nil {
		return Team{}, err
	}
	return newTeam(m.deps, team), nil
}

func newTeam(deps *Dependencies, team entity.Team) Team {
	return Team{
		deps: deps,
		Team: team,
	}
}

func (m Mutation) Team(ctx context.Context,
	args struct{ ID graphql.ID },
) (TeamUpdate, error) {
	ts := m.deps.Data.FilterTeams(func(t entity.Team) bool {
		return toGraphQLID(t.ID) == args.ID
	})
	if len(ts) == 0 {
		return TeamUpdate{}, errors.Errorf("team %v is not found", args.ID)
	}
	return TeamUpdate{team: ts[0], deps: m.deps}, nil
}

type TeamUpdate struct {
	team entity.Team
	deps *Dependencies
}

func (tu TeamUpdate) AddMember(args struct{ ID graphql.ID }) (TeamUpdate, error) {
	userID, err := fromGraphQLID(args.ID)
	if err != nil {
		return TeamUpdate{}, err
	}
	newTeam, err := tu.deps.Data.UpdateTeam(tu.team.ID, func(t entity.Team) entity.Team {
		t.MemberIDs = t.MemberIDs.Add(userID)
		return t
	})
	if err != nil {
		return TeamUpdate{}, err
	}
	tu.team = newTeam
	return tu, nil
}

func (m TeamUpdate) PromoteTaskToNeedAttention(
	ctx context.Context,
	args struct {
		ID graphql.ID
	},
) (TeamUpdate, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		return TeamUpdate{}, err
	}
	user, err := m.deps.Data.GetUser(userID)
	if err != nil {
		return TeamUpdate{}, err
	}
	taskID, err := fromGraphQLID(args.ID)
	if err != nil {
		return TeamUpdate{}, err
	}
	task, err := m.deps.Data.GetTask(args.ID)
	if err != nil {
		return TeamUpdate{}, err
	}
	if task.Status != entity.IN_PROGRESS {
		return TeamUpdate{}, errors.Errorf("task %v is not in progress, can not promote to Need Attention", taskID)
	}
	team, err := m.deps.Data.UpdateTeam(user.ActiveTeamID, func(t entity.Team) entity.Team {
		t.NeedAttentionTasks[userID] = taskID
		return t
	})
	if err != nil {
		return TeamUpdate{}, err
	}
	return TeamUpdate{deps: m.deps, team: team}, nil
}

func (tu TeamUpdate) Team() Team {
	return newTeam(tu.deps, tu.team)
}

func newTeams(deps *Dependencies, teams []entity.Team) (ts []Team) {
	for _, t := range teams {
		ts = append(ts, newTeam(deps, t))
	}
	return
}
