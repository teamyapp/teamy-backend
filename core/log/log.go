package log

import (
	"context"
	"log"
	"strings"

	"github.com/teamyapp/cloud/app/ctx"
)

func Info(args ...interface{}) {
	if len(args) == 0 {
		return
	}
	first := args[0]
	if ct, ok := first.(context.Context); ok {
		logs := args[1:]
		reqID, ok := ct.Value("request-id").(string)
		if ok {
			logs = append([]interface{}{reqID}, logs...)
		}
		userID, err := ctx.UserIDFromContext(ct)
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
