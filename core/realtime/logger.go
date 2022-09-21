package realtime

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
)

type MutationLogger struct {
	logger obs.Logger
}

var _ obs.Logger = (*MutationLogger)(nil)

func (m MutationLogger) Log(severity obs.Severity, props obs.Props) {
	m.LogAndSkip(severity, props, 1)
}

func (m MutationLogger) LogAndSkip(severity obs.Severity, props obs.Props, skipCallers int) {
	m.logger.LogAndSkip(severity, props, skipCallers+1)
}

func (m MutationLogger) LogWithContext(ct context.Context, severity obs.Severity, props obs.Props) {
	m.LogWithContextAndSkip(ct, severity, props, 1)
}

func (m MutationLogger) LogWithContextAndSkip(ct context.Context, severity obs.Severity, props obs.Props, skipCallers int) {
	newProps := obs.Props{}
	for key, value := range props {
		newProps[key] = value
	}

	mutationID, ok := GetMutationID(ct)
	if ok {
		newProps["mutationId"] = mutationID
	}

	m.logger.LogWithContextAndSkip(ct, severity, newProps, skipCallers+1)
}

func NewMutationLogger(logger obs.Logger) MutationLogger {
	return MutationLogger{
		logger: logger,
	}
}
