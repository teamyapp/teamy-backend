package resolver

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/one/identity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Query struct {
	deps *Dependencies
}

func (q Query) Task(args struct {
	ID graphql.ID
}) (Task, error) {
	task, err := q.deps.Data.GetTask(args.ID)
	if err != nil {
		return Task{}, err
	}
	return newTask(q.deps, task), nil
}

func (q Query) Tasks(args struct{ Input *TaskFilter }) ([]Task, error) {
	tasks := q.deps.Data.FilterTasks(func(t entity.Task) bool {
		return taskFilterFunc(t, args.Input)
	})
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
	return newTasks(q.deps, tasks), nil
}

func (q Query) Me(ctx context.Context) (User, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	user, err := q.deps.Data.GetUser(userID)
	if err != nil {
		log.Printf("%+v\n", err)
		return User{}, err
	}
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	return newUser(q.deps, user), nil
}

// debug only
func (q Query) Teams(ctx context.Context, args struct {
	IDs *[]graphql.ID
}) ([]Team, error) {
	var teams []entity.Team

	if args.IDs == nil {
		teams = q.deps.Data.FilterTeams(func(team entity.Team) bool { return true })
	} else {
		idsMap, err := toIDsMap(*args.IDs)
		if err != nil {
			return nil, err
		}

		teams = q.deps.Data.FilterTeams(func(team entity.Team) bool {
			_, ok := idsMap[team.ID]
			return ok
		})
	}

	return newTeams(q.deps, teams), nil
}

func (q Query) Invitations(ctx context.Context, args struct {
	Input struct {
		ID   *graphql.ID
		Code *string
	}
}) ([]Invitation, error) {
	if args.Input.ID != nil {
		id, err := fromGraphQLID(*args.Input.ID)
		if err != nil {
			return nil, err
		}
		invitations := q.deps.Data.FilterInvitations(func(invitation entity.Invitation) bool {
			return invitation.ID == id
		})
		if len(invitations) != 1 {
			return nil, fmt.Errorf("must find only 1 invitation: id=%v", id)
		}
		invitation := invitations[0]

		userID, err := identity.FromContext(ctx)
		if err != nil {
			return nil, err
		}

		teams := q.deps.Data.FilterTeams(func(team entity.Team) bool {
			return team.ID == invitation.TeamID && team.CreatorID == userID
		})
		if len(teams) != 1 {
			return nil, fmt.Errorf("must find only 1 team: %v", id)
		}

		return []Invitation{{deps: q.deps, Invitation: invitation}}, nil
	}

	if args.Input.Code != nil {
		invitations := q.deps.Data.FilterInvitations(func(invitation entity.Invitation) bool {
			return invitation.Code == *args.Input.Code
		})
		if len(invitations) != 1 {
			return nil, fmt.Errorf("must find only 1 invitation: code=%v", args.Input.Code)
		}
		return []Invitation{{deps: q.deps, Invitation: invitations[0]}}, nil
	}

	invitations := q.deps.Data.FilterInvitations(func(invitation entity.Invitation) bool {
		return true
	})

	gqlInvitations := make([]Invitation, 0)
	for _, invitation := range invitations {
		gqlInvitations = append(gqlInvitations, Invitation{deps: q.deps, Invitation: invitation})
	}
	return gqlInvitations, nil
}

func NewQuery(deps *Dependencies) Query {
	return Query{
		deps: deps,
	}
}
