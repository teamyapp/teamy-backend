package realtimev2

import (
	"context"

	"github.com/teamyapp/cloud/libs/telemetry"
)

const MutationIDProp = "MutationId"

func MutationLogInterceptor(ct context.Context, level telemetry.LogLevel, props telemetry.Props) telemetry.Props {
	newProps := telemetry.Props{}
	for key, value := range props {
		newProps[key] = value
	}

	mutationID, ok := GetMutationID(ct)
	if ok {
		newProps[MutationIDProp] = mutationID
	}

	return newProps
}

var _ telemetry.LogInterceptor = MutationLogInterceptor
