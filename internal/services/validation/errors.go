package validationservice

type ErrValidationUserData struct {
	err string
}

func NewErrValidationUserData(err string) ErrValidationUserData {
	return ErrValidationUserData{
		err: err,
	}
}

func (e ErrValidationUserData) Error() string {
	return e.err
}
