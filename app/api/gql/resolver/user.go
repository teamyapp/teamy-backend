package resolver

import (
	"fmt"
	"log"

	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type User struct {
	Entity
	deps *Dependencies
	user entity.User
}

func (u User) FirstName() string {
	// TODO: replace with real first name
	return u.user.FirstName
}

func (u User) LastName() string {
	// TODO: replace with real last name
	return u.user.LastName
}

func (u User) ProfileURL() string {
	return u.user.ProfileURL
}

func (u User) ActiveTeam() (*Team, error) {
	teams := u.deps.Data.FilterTeams(func(t entity.Team) bool {
		return t.ID == u.user.ActiveTeamID
	})
	if len(teams) == 0 {
		return nil, nil
	}
	gqlTeam := newTeam(u.deps, teams[0])
	return &gqlTeam, nil
}

func (u User) Teams() ([]Team, error) {
	teams := u.deps.Data.FilterTeams(func(t entity.Team) bool {
		isCreator := t.CreatorID == u.user.ID
		if isCreator {
			return true
		}
		for _, id := range t.MemberIDs {
			if id == u.user.ID { // is member
				return true
			}
		}
		return false
	})
	return newTeams(u.deps, teams), nil
}

func (u User) Tasks(args struct{ Input *TaskFilter }) ([]Task, error) {
	if args.Input == nil {
		tasks := u.deps.Data.FilterTasks(func(t entity.Task) bool {
			ownerID := oneEntity.ID(-1)
			if t.OwnerUserId != nil {
				ownerID = *t.OwnerUserId
			}
			return t.CreatorID == u.ID() || ownerID == u.user.ID
		})
		return newTasks(u.deps, tasks), nil
	}
	if args.Input.Status == nil {
		return nil, nil
	}
	status := *(args.Input.Status)
	activeTeam, err := u.ActiveTeam()
	if err != nil {
		return nil, err
	}
	if activeTeam == nil {
		return nil, nil
	}
	switch status {
	case UPCOMING:
		upcomingTasks, err := u.deps.taskRepo.FindTasksForUser(u.entity.ID, activeTeam.entity.ID, entity.TaskStatusUpcoming)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		upcomingTasks = u.deps.prioritizationService.PrioritizeTasks(upcomingTasks)
		upcomingTasks = tasksWithAvailableActions(upcomingTasks, entity.TaskStatusUpcoming)
		return toGraphQLTasks(u.deps, upcomingTasks), nil
	case IN_PROGRESS:
		// not applicable to users for now
		// todo: might implement in the future
		return nil, nil
	case DELIVERED:
		deliveredTasks, err := u.deps.taskRepo.FindTasksForUser(u.entity.ID, activeTeam.entity.ID, entity.TaskStatusDelivered)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		deliveredTasks = tasksWithAvailableActions(deliveredTasks, entity.TaskStatusDelivered)
		return toGraphQLTasks(u.deps, deliveredTasks), nil
	default:
		return nil, fmt.Errorf("status %v does not exist", status)
	}
}

func newUser(deps *Dependencies, user entity.User) User {
	return User{
		Entity: Entity{entity: user.Entity},
		deps:   deps,
		user:   user,
	}
}
