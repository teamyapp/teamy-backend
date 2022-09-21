package realtime

import (
	"context"
)

const mutationIDKey = "T-Mutation-Id"

func GetMutationID(ctx context.Context) (uint64, bool) {
	value, ok := ctx.Value(mutationIDKey).(uint64)
	return value, ok
}

func WithMutationID(ctx context.Context, mutationID uint64) context.Context {
	return context.WithValue(ctx, mutationIDKey, mutationID)
}
