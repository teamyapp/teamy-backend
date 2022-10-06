package scalar

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/graph-gophers/graphql-go/decode"
	"github.com/teamyapp/cloud/libs/duration"
	"github.com/teamyapp/cloud/libs/obs"
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
	ct := context.Background()
	dataCollector := inject.Injector.Get(new(obs.DataCollector)).(obs.DataCollector)
	switch input.(type) {
	case string:
		var err error
		d.Duration, err = duration.Parse(ct, dataCollector, input.(string))
		if err != nil {
			dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return err
		}

	default:
		err := errors.New("unsupported duration dataType")
		dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
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
