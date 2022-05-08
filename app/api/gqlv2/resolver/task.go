package resolver

import (
	"context"

	"github.com/teamyapp/teamy-backend/app/entityv2"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Task struct {
	deps Dependencies
	task entityv2.Task
}

func (t Task) ID(ctx context.Context) (graphql.ID, error) {
	panic("implement me")
}

func (t Task) Goal(ctx context.Context) (string, error) {
	panic("implement me")
}

func (t Task) Context(ctx context.Context) (*string, error) {
	panic("implement me")
}

func (t Task) Creator(ctx context.Context) (User, error) {
	panic("implement me")
}

func (t Task) Owner(ctx context.Context) (*User, error) {
	panic("implement me")
}

func (t Task) OwningTeam(ctx context.Context) (*Team, error) {
	panic("implement me")
}

func (t Task) Status(ctx context.Context) (entity.TaskStatus, error) {
	panic("implement me")
}

func (t Task) Comments(ctx context.Context) (Thread, error) {
	panic("implement me")
}

func (t Task) CreatedAt(ctx context.Context) (graphql.Time, error) {
	panic("implement me")
}

func (t Task) UpdatedAt(ctx context.Context) (*graphql.Time, error) {
	panic("implement me")
}

func (t Task) DueAt(ctx context.Context) (*graphql.Time, error) {
	panic("implement me")
}

func (t Task) AvailableActions(ctx context.Context) ([]entity.TaskAction, error) {
	panic("implement me")
}
