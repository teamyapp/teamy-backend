package gqlv2

import _ "embed"

//go:embed schema.gql
var rawSchema string

func RawSchema() string {
	return rawSchema
}
