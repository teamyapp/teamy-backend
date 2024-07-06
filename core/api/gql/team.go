package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Team struct {
	deps *Dependencies
	team entity.Team
}

func (t Team) ID(ct context.Context) graphql.ID {
	return toGraphQLID(t.team.ID)
}

func (t Team) Name(ct context.Context) string {
	return t.team.Name
}

func (t Team) IconURL(ct context.Context) *string {
	return t.team.IconURL
}

func (t Team) CreatedAt(ct context.Context) graphql.Time {
	return toGraphQLTime(t.team.CreatedAt)
}

func (t Team) MaxGroupOrderIndex(ct context.Context) int32 {
	return int32(t.team.MaxGroupOrderIndex)
}

func (t Team) Creator(ct context.Context) (User, error) {
	user, err := t.deps.userService.FindUserByID(ct, t.team.CreatorUserID)
	if err != nil {
		t.deps.logger.ErrorWithContext(ct, err)
		return User{}, errs.ToResolverErr(err)
	}

	return newUser(t.deps, user), nil
}

func (t Team) Owner(ct context.Context) (User, error) {
	user, err := t.deps.userService.FindUserByID(ct, t.team.OwnerUserID)
	if err != nil {
		t.deps.logger.ErrorWithContext(ct, err)
		return User{}, nil
	}

	return newUser(t.deps, user), nil
}

func (t Team) ActiveSprint(ct context.Context) (*Sprint, error) {
	sprint, err := t.deps.sprintService.GetActiveSprint(ct, t.team.ID)
	if err != nil {
		t.deps.logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	if sprint == nil {
		return nil, nil
	}

	gqlSprint := newSprint(t.deps, *sprint)
	return &gqlSprint, nil
}

func (t Team) Members(ct context.Context) ([]TeamMember, error) {
	teamMembers, err := t.deps.teamService.FindTeamMembers(ct, t.team.ID)
	if err != nil {
		t.deps.logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(teamMembers, func(teamMember entity.TeamMember, _ int) TeamMember {
		return newTeamMember(t.deps, teamMember)
	}), nil
}

func (t Team) TaskActivities(ct context.Context) ([]TaskActivity, error) {
	taskActivities := t.deps.taskService.FindTaskActivities(ct, t.team.ID)
	return collect.Map(taskActivities, func(taskActivity entity.TaskActivity, _ int) TaskActivity {
		return newTaskActivity(t.deps, taskActivity)
	}), nil
}

func (t Team) Projects(ct context.Context) ([]Project, error) {
	projects, err := t.deps.projectService.FindProjectsByTeamID(ct, t.team.ID)
	if err != nil {
		t.deps.logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(projects, func(project entity.Project, _ int) Project {
		return newProject(t.deps, project)
	}), nil
}

func (t Team) Tasks(ct context.Context, args struct {
	Filter *TaskFilter
	Sort   *struct {
		Field TaskSortField
		Order SortOrder
	}
	Pagination *struct {
		PageSize    int32
		AfterCursor *string
	}
}) (Tasks, error) {
	filter, argErr := fromGraphQLTaskFilterPtr(args.Filter)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		t.deps.logger.ErrorWithContext(ct, internalErr)
		return nil, errs.ToResolverErr(internalErr)
	}

	tasks, err := t.deps.taskService.FindTasksInTeam(ct, t.team.ID, filter)
	if err != nil {
		t.deps.logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(tasks, func(task entity.Task, _ int) Task {
		return newTask(t.deps, task)
	}), nil
}

func (t Team) Invitations(ct context.Context, args struct {
	Filter *InvitationFilter
}) ([]Invitation, error) {
	filter, argErr := fromGraphQLInvitationFilterPtr(args.Filter)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		t.deps.logger.ErrorWithContext(ct, internalErr)
		return nil, errs.ToResolverErr(internalErr)
	}

	invitations, err := t.deps.invitationService.FindInvitationsInTeam(ct, t.team.ID, filter)
	if err != nil {
		t.deps.logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(invitations, func(invitationEntity entity.Invitation, _ int) Invitation {
		return newInvitation(t.deps, invitationEntity)
	}), nil
}

func (t Team) Sprints(ct context.Context, args struct {
	Filter *SprintFilter
}) ([]Sprint, error) {
	filter, argErr := fromGraphQLSprintFilterPtr(args.Filter)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		t.deps.logger.ErrorWithContext(ct, internalErr)
		return nil, errs.ToResolverErr(internalErr)
	}

	sprints, err := t.deps.sprintService.FindSprintsInTeam(ct, t.team.ID, filter)
	if err != nil {
		t.deps.logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(sprints, func(sprint entity.Sprint, index int) Sprint {
		return newSprint(t.deps, sprint)
	}), nil
}

func (t Team) AppInstallations(ct context.Context) ([]TeamAppInstallation, error) {
	teamAppInstallations, err := t.deps.appService.FindTeamAppInstallationsByTeamID(ct, t.team.ID)
	if err != nil {
		t.deps.logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(teamAppInstallations, func(teamAppInstallation entity.TeamAppInstallation, index int) TeamAppInstallation {
		return newTeamAppInstallation(t.deps, teamAppInstallation)
	}), nil
}

func (t Team) ManagedApps(ct context.Context) ([]App, error) {
	apps, err := t.deps.appService.FindAppsByManagedByTeamID(ct, t.team.ID)
	if err != nil {
		t.deps.logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(apps, func(app entity.App, index int) App {
		return newApp(t.deps, app)
	}), nil
}

func (t Team) MemberGroups(ct context.Context) ([]TeamMemberGroup, error) {
	teamMemberGroups, err := t.deps.teamService.FindTeamMemberGroups(ct, t.team.ID)
	if err != nil {
		t.deps.logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(teamMemberGroups, func(teamMemberGroup entity.TeamMemberGroup, index int) TeamMemberGroup {
		return newTeamMemberGroup(t.deps, teamMemberGroup)
	}), nil
}

func newTeam(deps *Dependencies, team entity.Team) Team {
	return Team{
		deps: deps,
		team: team,
	}
}
