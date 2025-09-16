package filesrepo

import "errors"

var (
	ErrFileNotFound = errors.New("file not found")
)

type ErrFileUpload struct {
	err string
}

func NewErrFileUpload(err string) ErrFileUpload {
	return ErrFileUpload{
		err: err,
	}
}

func (e ErrFileUpload) Error() string {
	return e.err
}