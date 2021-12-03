package resolver

import (
	"fmt"
	"log"

	"github.com/teamyapp/teamy-backend/app/api/gqlv2/resolver"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type User struct {
	Entity
	deps          *Dependencies
	prototypeDeps *resolver.Dependencies
	user          entity.User
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
	team, err := u.deps.teamRepo.FindActiveTeam(u.user.ID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if team == nil {
		return nil, nil
	}

	gqlTeam := newTeam(u.deps, u.prototypeDeps, *team)
	return &gqlTeam, nil
}

func (u User) Teams() ([]Team, error) {
	teamIDs, err := u.deps.teamRepo.FindAllTeamIDs(u.user.ID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	teams, err := u.deps.teamRepo.FindTeams(teamIDs)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var gqlTeams = make([]Team, 0)
	for _, team := range teams {
		gqlTeam := newTeam(u.deps, u.prototypeDeps, team)
		gqlTeams = append(gqlTeams, gqlTeam)
	}
	return gqlTeams, nil
}

func (u User) Tasks(args struct{ Input *TaskFilter }) ([]Task, error) {
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
		return toGraphQLTasks(u.deps, u.prototypeDeps, upcomingTasks), nil
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
		return toGraphQLTasks(u.deps, u.prototypeDeps, deliveredTasks), nil
	default:
		return nil, fmt.Errorf("status %v does not exist", status)
	}
}

func newUser(deps *Dependencies, prototypeDeps *resolver.Dependencies, user entity.User) User {
	return User{
		Entity:        Entity{entity: user.Entity},
		deps:          deps,
		prototypeDeps: prototypeDeps,
		user:          user,
	}
}
