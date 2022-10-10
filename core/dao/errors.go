package dao

type ErrNotFound string

var ErrorNotFound ErrNotFound = ""

func (e ErrNotFound) Error() string {
	return string(e)
}
