package failure

import "errors"

type InternalError struct {
	baseError
}

func NewInternalError(err error) error {
	return InternalError{
		baseError: newBaseError(err),
	}
}

func (err InternalError) Error() string {
	return "internal error: " + err.baseError.Error()
}

func IsInternalError(err error) bool {
	return errors.As(err, new(InternalError))
}
