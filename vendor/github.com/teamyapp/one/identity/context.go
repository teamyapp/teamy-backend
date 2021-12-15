package identity

import (
	"context"
	"fmt"

	"github.com/teamyapp/one/entity"
)

type key int

const userIDKey key = 0

func FromContext(ctx context.Context) (entity.ID, error) {
	userID, ok := ctx.Value(userIDKey).(entity.ID)
	if !ok {
		return 0, fmt.Errorf("userID not found")
	}

	return userID, nil
}

func newContext(ctx context.Context, userID entity.ID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
