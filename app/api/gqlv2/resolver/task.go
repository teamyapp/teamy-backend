package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Task struct {
}

func (t Task) ID(ctx context.Context) (graphql.ID, error) {
}

func (t Task) Goal(ctx context.Context) (string, error) {
}

func (t Task) Context(ctx context.Context) (*string, error) {
}

func (t Task) Creator(ctx context.Context) (User, error) {
}

func (t Task) Owner(ctx context.Context) (*User, error) {
}

func (t Task) OwningTeam(ctx context.Context) (*Team, error) {
}

func (t Task) Status(ctx context.Context) (entity.TaskStatus, error) {
}

func (t Task) Comments(ctx context.Context) (Thread, error) {
}

func (t Task) CreatedAt(ctx context.Context) (graphql.Time, error) {
}

func (t Task) UpdatedAt(ctx context.Context) (*graphql.Time, error) {
}

func (t Task) DueAt(ctx context.Context) (*graphql.Time, error) {
}

func (t Task) AvailableActions(ctx context.Context) ([]entity.TaskAction, error) {
}
