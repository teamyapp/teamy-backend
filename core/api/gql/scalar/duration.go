package scalar

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/graph-gophers/graphql-go/decode"
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
		du, err := time.ParseDuration(input.(string))
		if err != nil {
			return err
		}

		d.Duration = du
	default:
		return fmt.Errorf("unsupported duration format: %v", input)
	}

	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}
