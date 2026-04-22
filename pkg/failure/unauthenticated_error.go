package failure

import (
	"errors"
)

type UnauthenticatedError struct {
	baseError
}

func NewUnauthenticatedError(err error) error {
	return UnauthenticatedError{
		baseError: newBaseError(err),
	}
}

func (err UnauthenticatedError) Error() string {
	return "unauthenticated: " + err.baseError.Error()
}

func IsNotUnauthenticatedError(err error) bool {
	return errors.As(err, new(UnauthenticatedError))
}
