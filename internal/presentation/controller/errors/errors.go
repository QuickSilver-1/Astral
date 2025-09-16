package controllererrors

type ErrInvalidInputData struct {
	err string
}

func NewErrInvalidInputData(err string) ErrInvalidInputData {
	return ErrInvalidInputData{
		err: err,
	}
}

func (e ErrInvalidInputData) Error() string {
	return "invalid input data: " + e.err
}