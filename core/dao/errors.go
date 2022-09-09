package dao

type ErrNotFound string

var ErrorNotFound = (*ErrNotFound)(nil)

func (e ErrNotFound) Error() string {
	return string(e)
}
