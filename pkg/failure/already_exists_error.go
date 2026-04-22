package failure

import "errors"

type AlreadyExistsError struct {
	baseError
}

func NewAlreadyExistsError(err error) error {
	return AlreadyExistsError{
		baseError: newBaseError(err),
	}
}

func (err AlreadyExistsError) Error() string {
	return "already exist error: " + err.baseError.Error()
}

func IsAlreadyExistsError(err error) bool {
	return errors.As(err, new(AlreadyExistsError))
}
