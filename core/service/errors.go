package service

type ErrNotFound string

var ErrorNotFound *ErrNotFound
var _ error = (*ErrNotFound)(nil)

func (e ErrNotFound) Error() string {
	return string(e)
}
