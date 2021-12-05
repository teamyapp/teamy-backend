package resolver

import (
	"context"
	"log"
	"strings"

	"github.com/graph-gophers/graphql-go"
	oneEntity "github.com/teamyapp/one/entity"
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
		if args.Input == nil {
			return true
		}
		// filter by Creator
		matchCreator := true
		if args.Input.CreatorID != nil {
			creatorID := *args.Input.CreatorID
			ownerID := oneEntity.ID(-1)
			if t.OwnerUserId != nil {
				ownerID = *t.OwnerUserId
			}
			matchCreator = t.CreatorID == creatorID || toGraphQLID(ownerID) == creatorID
			log.Println(matchCreator)
		}
		// filter by status
		matchStatus := true
		if args.Input.Status != nil {
			status := *(args.Input.Status)
			matchStatus = t.Status == status
			log.Println("matchStatus", matchStatus)
		}
		// full text search
		// todo: need to implement a better full text search
		// by using the full-text search engine in postgres
		matchText := true
		if args.Input.Text != nil {
			text := *(args.Input.Text)
			matchGoal := strings.Contains(t.Goal, text)
			matchContext := strings.Contains(t.Context, text)
			matchText = matchGoal || matchContext
			log.Println(matchText)
		}
		log.Println("=", matchCreator, matchStatus, matchText)
		return matchCreator && matchStatus && matchText
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
func (q Query) Teams(ctx context.Context) ([]Team, error) {
	teams := q.deps.Data.FilterTeams(func(t entity.Team) bool { return true })
	return newTeams(q.deps, teams), nil
}

func NewQuery(deps *Dependencies) Query {
	return Query{
		deps: deps,
	}
}
