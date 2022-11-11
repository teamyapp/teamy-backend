package realtime

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
)

type MutationLogger struct {
	logger obs.Logger
}

var _ obs.Logger = (*MutationLogger)(nil)

func (m MutationLogger) Log(level obs.LogLevel, props obs.Props) {
	m.LogAndSkip(level, props, 1)
}

func (m MutationLogger) LogAndSkip(level obs.LogLevel, props obs.Props, skipCallers int) {
	m.logger.LogAndSkip(level, props, skipCallers+1)
}

func (m MutationLogger) LogWithContext(ct context.Context, level obs.LogLevel, props obs.Props) {
	m.LogWithContextAndSkip(ct, level, props, 1)
}

func (m MutationLogger) LogWithContextAndSkip(ct context.Context, level obs.LogLevel, props obs.Props, skipCallers int) {
	newProps := obs.Props{}
	for key, value := range props {
		newProps[key] = value
	}

	mutationID, ok := GetMutationID(ct)
	if ok {
		newProps["MutationId"] = mutationID
	}

	m.logger.LogWithContextAndSkip(ct, level, newProps, skipCallers+1)
}

func NewMutationLogger(logger obs.Logger) MutationLogger {
	return MutationLogger{
		logger: logger,
	}
}
