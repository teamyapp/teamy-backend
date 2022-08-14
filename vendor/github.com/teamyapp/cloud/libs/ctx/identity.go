package ctx

import (
	"context"
	"fmt"
	"log"
)

func UserIDFromContext(ctx context.Context) (uint64, error) {
	userID, ok := ctx.Value(userIDKey).(uint64)
	if !ok {
		err := fmt.Errorf("userID not found")
		log.Println(err)
		return 0, err
	}

	return userID, nil
}

func NewContextWithUserID(ctx context.Context, userID uint64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
