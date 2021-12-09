package operations

import _ "embed"

//go:embed operations.gql
var operations string

func Operations() string {
	return operations
}
