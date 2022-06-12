package dao

type ErrNotFound string

var _ error = (*ErrNotFound)(nil)

func (e ErrNotFound) Error() string {
	return string(e)
}
