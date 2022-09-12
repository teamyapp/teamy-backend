package gql

import "fmt"

type ErrorCode string

const (
	unauthorizedErrorCode ErrorCode = "Unauthorized"
)

type ResolverError struct {
	Code    ErrorCode
	Message string
}

var _ error = (*ResolverError)(nil)

func (r ResolverError) Error() string {
	return fmt.Sprintf("code=%s, message=%s", r.Code, r.Message)
}

func (r ResolverError) Extensions() map[string]interface{} {
	return map[string]interface{}{
		"code":    r.Code,
		"message": r.Message,
	}
}
