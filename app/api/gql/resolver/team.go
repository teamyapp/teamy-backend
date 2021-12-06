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
	// TODO: (Begin) remove once JSON data feed is ready
	memberIDs, err := t.deps.teamRepo.ListTeamMemberIDs(t.entity.ID)
	if err != nil {
		return nil, err
	}
	members, err := t.deps.userRepo.FindUsers(memberIDs)
	if err != nil {
		return nil, err
	}
	// TODO: (End) remove once JSON data feed is ready

	// TODO: (Begin) enable once JSON data feed is ready
	//members, err := t.deps.Data.GetUsers(t.team.MemberIDs)
	// TODO: (End) enable once JSON data feed is ready
	return toGraphQLUsers(t.deps, members), err
}

func (t Team) Tasks(args struct{ Input *TaskFilter }) ([]Task, error) {
	// TODO: (Begin) enable once JSON data feed is ready
	//teams := t.deps.Data.FilterTeams(func(team entity.Team) bool {
	//	return team.ID == t.entity.ID
	//})
	//if len(teams) != 1 {
	//	return nil, fmt.Errorf("more than 1 team found for %v", t.entity.ID)
	//}
	// TODO: (End) enable once JSON data feed is ready

	if args.Input == nil {
		// TODO: return all related tasks
		return nil, nil
	}
	if args.Input.Status == nil {
		return nil, nil
	}
	status := *(args.Input.Status)
	switch status {
	case upcoming:
		upcomingTasks, err := t.deps.taskRepo.FindTasksForTeam(t.entity.ID, entity.TaskStatusUpcoming)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		upcomingTasks = t.deps.prioritizationService.PrioritizeTasks(upcomingTasks)
		upcomingTasks = tasksWithAvailableActions(upcomingTasks, entity.TaskStatusUpcoming)
		return toGraphQLTasks(t.deps, upcomingTasks), nil
	case inProgress:
		upcomingTasks, err := t.deps.taskRepo.FindTasksForTeam(t.entity.ID, entity.TaskStatusInProgress)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		upcomingTasks = t.deps.prioritizationService.PrioritizeTasks(upcomingTasks)
		upcomingTasks = tasksWithAvailableActions(upcomingTasks, entity.TaskStatusInProgress)
		return toGraphQLTasks(t.deps, upcomingTasks), nil
	case delivered:
		deliveredTasks, err := t.deps.taskRepo.FindTasksForTeam(t.entity.ID, entity.TaskStatusDelivered)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		deliveredTasks = tasksWithAvailableActions(deliveredTasks, entity.TaskStatusDelivered)
		return toGraphQLTasks(t.deps, deliveredTasks), nil
	default:
		return nil, fmt.Errorf("status %v does not exist", status)
	}
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
