package log

import (
	"context"
	"log"
	"strings"

	"github.com/teamyapp/one/identity"
)

func Info(args ...interface{}) {
	if len(args) == 0 {
		return
	}
	first := args[0]
	if ctx, ok := first.(context.Context); ok {
		logs := args[1:]
		reqID, ok := ctx.Value("request-id").(string)
		if ok {
			logs = append([]interface{}{reqID}, logs...)
		}
		userID, err := identity.FromContext(ctx)
		if err != nil && !strings.Contains(err.Error(), "userID not found") {
			log.Println(reqID, err)
		}
		if err == nil {
			logs = append(logs, "user", userID)
		}
		log.Println(logs...)
	} else {
		log.Println(args...)
	}
}
