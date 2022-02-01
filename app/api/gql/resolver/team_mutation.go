package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/one/identity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

func (m Mutation) CreateTeam(ctx context.Context,
	args struct {
		Input struct {
			Name string
		}
	},
) (TeamUpdate, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		return TeamUpdate{}, err
	}
	t, err := m.deps.Data.CreateTeam(userID, entity.Team{
		Name:      args.Input.Name,
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

func (m Mutation) Team(ctx context.Context, args struct{ ID graphql.ID },
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

func (tu TeamUpdate) Update(args struct {
	Input struct {
		Name    *string
		IconURL *string
	}
}) (TeamUpdate, error) {
	newTeam, err := tu.deps.Data.UpdateTeam(tu.team.ID, func(t entity.Team) entity.Team {
		if args.Input.Name != nil {
			t.Name = *args.Input.Name
		}
		if args.Input.IconURL != nil {
			t.IconURL = *args.Input.IconURL
		}
		return t
	})
	tu.team = newTeam
	return tu, err
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

func (tu TeamUpdate) RemoveMember(args struct{ ID graphql.ID }) (TeamUpdate, error) {
	userID, err := fromGraphQLID(args.ID)
	if err != nil {
		return TeamUpdate{}, err
	}
	newTeam, err := tu.deps.Data.UpdateTeam(tu.team.ID, func(t entity.Team) entity.Team {
		t.MemberIDs = t.MemberIDs.Remove(userID)
		return t
	})
	if err != nil {
		return TeamUpdate{}, err
	}
	tu.team = newTeam
	return tu, nil
}

func (tu TeamUpdate) RemoveTask(args struct{ ID graphql.ID }) (TeamUpdate, error) {
	taskID, err := fromGraphQLID(args.ID)
	if err != nil {
		return TeamUpdate{}, err
	}
	newTeam, err := tu.deps.Data.UpdateTeam(tu.team.ID, func(t entity.Team) entity.Team {
		t.Tasks = t.Tasks.Remove(taskID)
		return t
	})
	if err != nil {
		return TeamUpdate{}, err
	}
	tu.team = newTeam
	return tu, nil
}

func (tu TeamUpdate) PromoteTaskToNeedAttention(
	ctx context.Context,
	args struct {
		ID graphql.ID
	},
) (TeamUpdate, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		return TeamUpdate{}, err
	}
	user, err := tu.deps.Data.GetUser(userID)
	if err != nil {
		return TeamUpdate{}, err
	}
	taskID, err := fromGraphQLID(args.ID)
	if err != nil {
		return TeamUpdate{}, err
	}
	task, err := tu.deps.Data.GetTask(args.ID)
	if err != nil {
		return TeamUpdate{}, err
	}
	if task.Status != entity.IN_PROGRESS {
		return TeamUpdate{}, errors.Errorf("task %v is not in progress, can not promote to Need Attention", taskID)
	}
	team, err := tu.deps.Data.UpdateTeam(user.ActiveTeamID, func(t entity.Team) entity.Team {
		t.NeedAttentionTasks[userID] = taskID
		return t
	})
	if err != nil {
		return TeamUpdate{}, err
	}
	return TeamUpdate{deps: tu.deps, team: team}, nil
}

func (tu TeamUpdate) Team() Team {
	return newTeam(tu.deps, tu.team)
}
