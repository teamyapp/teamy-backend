package authorization

import "fmt"

type ErrorCode string

const (
	UnauthorizedErrorCode ErrorCode = "Unauthorized"
)

type Error struct {
	Code    ErrorCode
	Message string
}

var _ error = (*Error)(nil)

func (e Error) Error() string {
	return fmt.Sprintf("code=%s, message=%s", e.Code, e.Message)
}

func (e Error) Extensions() map[string]interface{} {
	return map[string]interface{}{
		"code":    e.Code,
		"message": e.Message,
	}
}
