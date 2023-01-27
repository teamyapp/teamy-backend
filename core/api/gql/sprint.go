package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/entity"
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
	filter, err := fromGraphQLTaskFilterPtr(ct, s.deps.dataCollector, args.Filter)
	if err != nil {
		s.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	tasks, err := s.deps.taskService.FindTasksInSprint(ct, s.sprint.ID, filter)
	if err != nil {
		s.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	return collect.Map(tasks, func(task entity.Task, index int) Task {
		return newTask(s.deps, task)
	}), nil
}

func (s Sprint) OwningTeam(ct context.Context) (Team, error) {
	team, err := s.deps.teamService.FindTeamByID(ct, s.sprint.OwningTeamID)
	if err != nil {
		s.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return Team{}, err
	}

	return newTeam(s.deps, team), nil
}

func (s Sprint) Participants(ct context.Context) ([]SprintParticipant, error) {
	participants, err := s.deps.sprintService.FindParticipantsInSprint(ct, s.sprint.ID)
	if err != nil {
		s.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	return collect.Map(participants, func(participant entity.SprintParticipant, index int) SprintParticipant {
		return newSprintParticipant(s.deps, participant)
	}), nil
}

func newSprint(deps *Dependencies, sprint entity.Sprint) Sprint {
	return Sprint{
		deps:   deps,
		sprint: sprint,
	}
}
