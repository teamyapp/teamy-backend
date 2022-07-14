package gql

import (
	"context"
	"fmt"
	"log"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/service"
)

type Sprint struct {
	deps   *Dependencies
	sprint entity.Sprint
}

func (s Sprint) ID(ct context.Context) graphql.ID {
	return toGraphQLID(s.sprint.ID)
}

func (s Sprint) StartAt(ct context.Context) graphql.Time {
	return toGraphQLTime(s.sprint.StartAt)
}

func (s Sprint) EndAt(ct context.Context) graphql.Time {
	return toGraphQLTime(s.sprint.EndAt)
}

func (s Sprint) CreatedAt(ct context.Context) graphql.Time {
	return toGraphQLTime(s.sprint.CreatedAt)
}

func (s Sprint) Tasks(ct context.Context, args struct {
	Filter *TaskFilter
}) ([]Task, error) {
	filter, err := fromGraphQLTaskFilterPtr(args.Filter)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tasks, err := s.deps.sprintService.FindTasksInSprint(ct, s.sprint.ID, filter)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return collect.Map(tasks, func(task entity.Task, index int) Task {
		return newTask(s.deps, task)
	}), nil
}

func (s Sprint) OwningTeam(ct context.Context) (Team, error) {
	filter := &service.TeamFilter{TeamID: &s.sprint.OwningTeamID}
	teams, err := s.deps.teamService.FindTeams(ct, filter)
	if err != nil {
		log.Println(err)
		return Team{}, err
	}

	if len(teams) == 0 {
		return Team{}, fmt.Errorf("team not found: teamID=%v", s.sprint.OwningTeamID)
	}

	return newTeam(s.deps, teams[0]), nil
}

func newSprint(deps *Dependencies, sprint entity.Sprint) Sprint {
	return Sprint{
		deps:   deps,
		sprint: sprint,
	}
}
