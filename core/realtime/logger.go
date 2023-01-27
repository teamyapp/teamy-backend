package realtime

import (
	"context"

	"github.com/teamyapp/cloud/libs/telemetry"
)

type MutationLogger struct {
	logger telemetry.Logger
}

var _ telemetry.Logger = (*MutationLogger)(nil)

func (m MutationLogger) Log(level telemetry.LogLevel, props telemetry.Props) {
	m.LogAndSkip(level, props, 1)
}

func (m MutationLogger) LogAndSkip(level telemetry.LogLevel, props telemetry.Props, skipCallers int) {
	m.logger.LogAndSkip(level, props, skipCallers+1)
}

func (m MutationLogger) LogWithContext(ct context.Context, level telemetry.LogLevel, props telemetry.Props) {
	m.LogWithContextAndSkip(ct, level, props, 1)
}

func (m MutationLogger) LogWithContextAndSkip(ct context.Context, level telemetry.LogLevel, props telemetry.Props, skipCallers int) {
	newProps := telemetry.Props{}
	for key, value := range props {
		newProps[key] = value
	}

	mutationID, ok := GetMutationID(ct)
	if ok {
		newProps["MutationId"] = mutationID
	}

	m.logger.LogWithContextAndSkip(ct, level, newProps, skipCallers+1)
}

func NewMutationLogger(logger telemetry.Logger) MutationLogger {
	return MutationLogger{
		logger: logger,
	}
}
