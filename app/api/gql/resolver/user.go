package resolver

import (
	"fmt"
	"log"

	"github.com/teamyapp/teamy-backend/app/entity"
)

type User struct {
	Entity
	deps *Dependencies
	user entity.User
}

func (u User) FirstName() string {
	return u.user.FirstName
}

func (u User) LastName() string {
	return u.user.LastName
}

func (u User) ProfileURL() string {
	return u.user.ProfileURL
}

func (u User) ActiveTeam() (*Team, error) {
	// TODO: (Begin) remove once JSON data feed is ready
	team, err := u.deps.teamRepo.FindActiveTeam(u.entity.ID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, nil
	}

	teams := []entity.Team{
		*team,
	}
	// TODO: (End) remove once JSON data feed is ready

	// TODO: (Begin) enable once JSON data feed is ready
	//teams := u.deps.Data.FilterTeams(func(t entity.Team) bool {
	//	return t.ID == u.user.ActiveTeamID
	//})
	//if len(teams) == 0 {
	//	return nil, nil
	//}
	// TODO: (End) enable once JSON data feed is ready
	gqlTeam := newTeam(u.deps, teams[0])
	return &gqlTeam, nil
}

func (u User) Teams() ([]Team, error) {
	// TODO: (Begin) remove once JSON data feed is ready
	teamIDs, err := u.deps.teamRepo.FindAllTeamIDs(u.entity.ID)
	if err != nil {
		return nil, err
	}
	teams, err := u.deps.teamRepo.FindTeams(teamIDs)
	// TODO: (End) remove once JSON data feed is ready

	// TODO: (Begin) enable once JSON data feed is ready
	//teams := u.deps.Data.FilterTeams(func(t entity.Team) bool {
	//	isCreator := t.CreatorID == u.user.ID
	//	if isCreator {
	//		return true
	//	}
	//	for _, id := range t.MemberIDs {
	//		if id == u.user.ID { // is member
	//			return true
	//		}
	//	}
	//	return false
	//})
	// TODO: (End) enable once JSON data feed is ready
	return newTeams(u.deps, teams), nil
}

func (u User) Tasks(args struct{ Input *TaskFilter }) ([]Task, error) {
	if args.Input == nil {
		// TODO: return all related tasks
		return nil, nil
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
	case upcoming:
		upcomingTasks, err := u.deps.taskRepo.FindTasksForUser(u.entity.ID, activeTeam.entity.ID, entity.TaskStatusUpcoming)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		upcomingTasks = u.deps.prioritizationService.PrioritizeTasks(upcomingTasks)
		upcomingTasks = tasksWithAvailableActions(upcomingTasks, entity.TaskStatusUpcoming)
		return toGraphQLTasks(u.deps, upcomingTasks), nil
	case inProgress:
		// not applicable to users for now
		// TODO: might implement in the future
		return nil, nil
	case delivered:
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

func (u User) NeedAttentionTask() (*Task, error) {
	activeTeam, err := u.ActiveTeam()
	if err != nil {
		return nil, err
	}
	if activeTeam == nil {
		return nil, nil
	}
	taskPtr, err := u.deps.taskRepo.FindTaskNeedAttentionForUser(u.entity.ID, activeTeam.entity.ID)
	if err != nil {
		return nil, err
	}
	if taskPtr == nil {
		return nil, nil
	}

	task := taskWithAvailableActions(*taskPtr, entity.TaskStatusNeedAttention)
	gqlTask := newTask(u.deps, task)
	return &gqlTask, nil
}

func newUser(deps *Dependencies, user entity.User) User {
	return User{
		Entity: Entity{entity: user.Entity},
		deps:   deps,
		user:   user,
	}
}
