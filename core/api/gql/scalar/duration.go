package scalar

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/graph-gophers/graphql-go/decode"
	"github.com/teamyapp/cloud/libs/duration"
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
	switch input.(type) {
	case string:
		var err error
		d.Duration, err = duration.Parse(input.(string))
		if err != nil {
			log.Println(err)
			return err
		}

	default:
		return fmt.Errorf("unsupported duration dataType: type=%v", reflect.TypeOf(input))
	}

	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(duration.Format(d.Duration))
}
