package ctx

import (
	"context"
	"fmt"
	"log"
)

func RequestIDFromContext(ctx context.Context) (string, error) {
	requestID, ok := ctx.Value(requestIDKey).(string)
	if !ok {
		err := fmt.Errorf("requestID not found")
		log.Println(err)
		return "", err
	}

	return requestID, nil
}

func NewContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}
