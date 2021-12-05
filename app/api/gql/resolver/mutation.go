package resolver

import (
	"context"
	"fmt"
	"log"

	"github.com/graph-gophers/graphql-go"
	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/one/identity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Mutation struct {
	deps  *Dependencies
	query *Query
}

func (m Mutation) CreateTask(
	ctx context.Context,
	args struct {
		Task TaskInput
	},
) (Task, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	// TODO: (Begin) remove once JSON data feed is ready
	task, err := fromGraphQLTaskInput(args.Task)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	activeTeam, err := m.deps.teamRepo.FindActiveTeam(userID)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}
	if activeTeam == nil {
		return Task{}, fmt.Errorf("user %v does not have an active team", userID)
	}

	taskID, err := m.deps.taskRepo.CreateTask(task)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	err = m.deps.taskRepo.AssignTaskToTeam(taskID, activeTeam.ID, entity.TaskStatusUpcoming)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	task, err = m.deps.taskRepo.FindTaskByID(taskID)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	if err != nil {
		log.Println(err)
		return Task{}, err
	}
	// TODO: (End) remove once JSON data feed is ready

	// TODO: (Begin) enable once JSON data feed is ready
	//// find user
	//user, err := m.deps.Data.GetUser(userID)
	//if err != nil {
	//	log.Println(err)
	//	return Task{}, err
	//}
	//
	//task, err := fromGraphQLTaskInput(args.Task)
	//if err != nil {
	//	log.Println(err)
	//	return Task{}, err
	//}
	//
	//activeTeams := m.deps.Data.FilterTeams(func(t entity.Team) bool {
	//	return t.ID == user.ActiveTeamID
	//})
	//if err != nil {
	//	log.Println(err)
	//	return Task{}, err
	//}
	//if len(activeTeams) == 0 {
	//	return Task{}, fmt.Errorf("user %v does not have an active team", userID)
	//}
	//task, err = m.deps.Data.CreateTask(toGraphQLID(userID), activeTeams[0].ID, task)
	//if err != nil {
	//	return Task{}, err
	//}
	//// add task to team
	//err = m.deps.Data.UpdateTeam(activeTeams[0].ID, func(t entity.Team) entity.Team {
	//	t.Tasks = append(t.Tasks, task.ID)
	//	return t
	//})
	//if err != nil {
	//	return Task{}, err
	//}
	// TODO: (End) enable once JSON data feed is ready
	return newTask(m.deps, task), err
}

func (m Mutation) StartTask(ctx context.Context, args struct {
	TaskID graphql.ID
}) (bool, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return false, err
	}

	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	// TODO: a user starts others' task will assign that task to the himself
	// TODO: show a modal to confirm task should be reassigned.
	activeTeam, err := m.deps.teamRepo.FindActiveTeam(userID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	prevNeedAttentionTaskID, err := m.deps.taskRepo.SetNeedAttentionTask(&taskID, userID, activeTeam.ID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	err = m.deps.taskRepo.SetTeamTaskStatus(taskID, activeTeam.ID, entity.TaskStatusInProgress)
	if err != nil {
		log.Println(err)
		return false, err
	}

	if prevNeedAttentionTaskID != nil {
		err = m.deps.taskRepo.SetTeamTaskStatus(*prevNeedAttentionTaskID, activeTeam.ID, entity.TaskStatusUpcoming)
		if err != nil {
			log.Println(err)
			return false, err
		}
	}

	return true, nil
}

func (m Mutation) DeleteTask(ctx context.Context, args struct {
	TaskID graphql.ID
}) (bool, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return false, err
	}

	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	// Check if taskID exists - no need, we can only delete task that we can see
	// UI:
	//		delete task from all active team view
	// Delete record from team_task table
	// TODO: move task to trash instead of completely deleting it. delete task after 7 days if in action
	// TODO: clean up the task from task dependency graph for the active team

	activeTeam, err := m.deps.teamRepo.FindActiveTeam(userID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	err = m.deps.taskRepo.DeleteTeamTask(taskID, activeTeam.ID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	err = m.deps.taskRepo.DeleteNeedAttentionTask(taskID, userID, activeTeam.ID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	return true, nil
}

func (m Mutation) CompleteTask(ctx context.Context, args struct {
	TaskID graphql.ID
}) (bool, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return false, err
	}

	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	activeTeam, err := m.deps.teamRepo.FindActiveTeam(userID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	needAttentionTask, err := m.deps.taskRepo.FindTaskNeedAttentionForUser(userID, activeTeam.ID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	if needAttentionTask.ID != taskID {
		return true, err
	}

	_, err = m.deps.taskRepo.SetNeedAttentionTask(nil, userID, activeTeam.ID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	err = m.deps.taskRepo.SetTeamTaskStatus(taskID, activeTeam.ID, entity.TaskStatusDelivered)
	if err != nil {
		log.Println(err)
		return false, err
	}

	return true, nil
}

func (m Mutation) UpdateActiveTeam(ctx context.Context, args struct {
	TeamID graphql.ID
}) (User, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	// TODO: (Begin) remove once JSON data feed is ready
	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		log.Printf("%+v\n", err)
		return User{}, err
	}

	_, err = m.deps.userRepo.FindUser(userID)
	if err != nil {
		return User{}, err
	}
	teamIDs, err := m.deps.teamRepo.FindAllTeamIDs(userID)
	if err != nil {
		return User{}, err
	}

	if !contains(teamIDs, teamID) {
		return User{}, fmt.Errorf("team not found: teamID=%v\n", teamID)
	}
	_, err = m.deps.userRepo.UpdateActiveTeamId(userID, &teamID)
	if err != nil {
		return User{}, err
	}
	user, err := m.deps.userRepo.FindUser(userID)
	if err != nil {
		log.Printf("%+v\n", err)
		return User{}, err
	}
	// TODO: (End) remove once JSON data feed is ready

	// TODO: (Begin) enable once JSON data feed is ready
	//user, err := m.deps.Data.GetUser(userID)
	//if err != nil {
	//	log.Printf("%+v\n", err)
	//	return User{}, err
	//}
	//id, err := fromGraphQLID(args.TeamID)
	//if err != nil {
	//	log.Printf("%+v\n", err)
	//	return User{}, err
	//}
	//user, err = m.deps.Data.UpdateUser(user.ID, func(u entity.User) entity.User {
	//	u.ActiveTeamID = id
	//	return u
	//})
	//if err != nil {
	//	log.Printf("%+v\n", err)
	//	return User{}, err
	//}
	// TODO: (End) enable once JSON data feed is ready
	return newUser(m.deps, user), err
}

func (m Mutation) UpdateTask(
	ctx context.Context,
	args struct {
		TaskID graphql.ID
		Task   TaskInput
	},
) (Task, error) {
	// TODO: need to do authorization
	// One can only update the task that is created by oneself.
	_, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}
	task, err := m.deps.Data.GetTask(args.TaskID)
	if err != nil {
		return Task{}, err
	}
	if args.Task.Context != nil {
		task.Context = args.Task.Context
	}
	if args.Task.DueAt != nil {
		task.DueAt = &args.Task.DueAt.Time
	}
	if args.Task.Goal != nil {
		task.Goal = *args.Task.Goal
	}
	if args.Task.OwnerUserID != nil {
		id, err := fromGraphQLIDPtr(args.Task.OwnerUserID)
		if err != nil {
			return Task{}, nil
		}
		task.OwnerUserId = id
	}
	task, err = m.deps.Data.UpdateTask(task)
	if err != nil {
		return Task{}, nil
	}
	return newTask(m.deps, task), nil
}

func (m Mutation) Comment(
	ctx context.Context,
	args struct {
		TaskID  graphql.ID
		Content string
	},
) (Comment, error) {
	// TODO: consider do this in a middleware and init User ID in the struct or ctx
	userID, err := identity.FromContext(ctx)
	if err != nil {
		return Comment{}, err
	}
	c, err := m.deps.Data.CreateComment(entity.Comment{
		Content:     args.Content,
		CommenterID: toGraphQLID(userID),
		TaskID:      args.TaskID,
	})
	if err != nil {
		return Comment{}, err
	}
	return Comment{
		deps:    m.deps,
		Comment: c,
	}, nil
}

func (m Mutation) CreateTeam(ctx context.Context,
	args struct {
		Input struct {
			Name string
		}
	},
) (Team, error) {
	// TODO: consider do this in a middleware and init User ID in the struct or ctx
	userID, err := identity.FromContext(ctx)
	if err != nil {
		return Team{}, err
	}
	t, err := m.deps.Data.CreateTeam(userID, entity.Team{
		Name:      args.Input.Name,
		CreatorID: userID,
		MemberIDs: []oneEntity.ID{
			userID,
		},
	})
	if err != nil {
		return Team{}, err
	}
	return newTeam(m.deps, t), nil
}

//
// Admin && Debug Only
//
func (m Mutation) CreateUser(args struct{ UserID graphql.ID }) (User, error) {
	id, err := fromGraphQLID(args.UserID)
	if err != nil {
		return User{}, err
	}
	user, err := m.deps.Data.CreateUser(id)
	if err != nil {
		return User{}, err
	}
	return newUser(m.deps, user), nil
}

func contains(arr []oneEntity.ID, element oneEntity.ID) bool {
	for _, e := range arr {
		if e == element {
			return true
		}
	}
	return false
}

func NewMutation(deps *Dependencies, query *Query) Mutation {
	return Mutation{
		deps:  deps,
		query: query,
	}
}
