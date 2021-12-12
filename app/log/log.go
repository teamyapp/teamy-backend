package log

import (
	"context"
	"log"
)

func Info(args ...interface{}) {
	if len(args) == 0 {
		return
	}
	first := args[0]
	if ctx, ok := first.(context.Context); ok {
		reqID, ok := ctx.Value("request-id").(string)
		if ok {
			log.Println(append([]interface{}{reqID}, args[1:]...)...)
		} else {
			log.Println(args[1:]...)
		}
	} else {
		log.Println(args...)
	}
}
