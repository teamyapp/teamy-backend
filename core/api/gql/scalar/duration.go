package scalar

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/graph-gophers/graphql-go/decode"
	"github.com/teamyapp/cloud/libs/duration"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/inject"
)

type Duration struct {
	time.Duration
}

var _ decode.Unmarshaler = (*Duration)(nil)
var _ json.Marshaler = (*Duration)(nil)

func (d Duration) ImplementsGraphQLType(name string) bool {
	return "Duration" == name
}

func (d *Duration) UnmarshalGraphQL(input interface{}) error {
	dataCollector := inject.Injector.Get(new(telemetry.DataCollector)).(telemetry.DataCollector)
	ct := context.Background()
	switch input.(type) {
	case string:
		var err error
		d.Duration, err = duration.Parse(ct, dataCollector, input.(string))
		if err != nil {
			dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			return err
		}

	default:
		err := errors.New("unsupported duration dataType")
		dataCollector.Logger.Log(telemetry.Error, telemetry.Props{
			telemetry.CauseProp: err,
			telemetry.MessageProp: telemetry.Props{
				"dataType": reflect.TypeOf(input),
			},
		})
		return err
	}

	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(duration.Format(d.Duration))
}
