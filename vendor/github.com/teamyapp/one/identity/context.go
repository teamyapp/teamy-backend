package identity

import (
	"context"
	"fmt"
	"log"

	"github.com/teamyapp/one/entity"
)

type key int

const userIDKey key = 0

func FromContext(ctx context.Context) (entity.ID, error) {
	userID, ok := ctx.Value(userIDKey).(entity.ID)
	if !ok {
		err := fmt.Errorf("userID not found")
		log.Println(err)
		return 0, err
	}

	return userID, nil
}

func newContext(ctx context.Context, userID entity.ID) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
