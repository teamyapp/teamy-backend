package lang

type Result[Value any] struct {
	Value Value
	Error error
}
