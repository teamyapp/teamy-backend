package scalar

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/graph-gophers/graphql-go/decode"
	"github.com/teamyapp/cloud/libs/duration"
	"github.com/teamyapp/cloud/libs/errs"
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
	logger := inject.Injector.Get(new(telemetry.Logger)).(telemetry.Logger)
	ct := context.Background()
	switch input.(type) {
	case string:
		var err *errs.Error
		d.Duration, err = duration.Parse(ct, input.(string))
		if err != nil {
			logger.ErrorWithContext(ct, err)
			return err.ToError()
		}

	default:
		err := &errs.Error{
			Code:    errs.InvalidArgument,
			Message: fmt.Sprintf("unsupported duration dataType: dataType=%v", reflect.TypeOf(input)),
		}
		logger.ErrorWithContext(ct, err)
		return err.ToError()
	}

	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(duration.Format(d.Duration))
}
